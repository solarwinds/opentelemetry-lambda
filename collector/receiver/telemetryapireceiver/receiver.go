// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/google/uuid"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/receiver"
	semconv "go.opentelemetry.io/collector/semconv/v1.25.0"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-lambda/collector/internal/telemetryapi"
)

const initialQueueSize = 5
const scopeName = "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapi"

type telemetryAPIReceiver struct {
	httpServer            *http.Server
	logger                *zap.Logger
	queue                 *queue.Queue // queue is a synchronous queue and is used to put the received log events to be dispatched later
	nextTraces            consumer.Traces
	nextMetrics           consumer.Metrics
	nextLogs              consumer.Logs
	lastPlatformStartTime string
	lastPlatformEndTime   string
	extensionID           string
	port                  int
	types                 []telemetryapi.EventType
	resource              pcommon.Resource
	metricsStartTime      time.Time
	coldStartCounter      int64
	errorsCounter         int64
	invocationsCounter    int64
	timeoutsCounter       int64
}

func (r *telemetryAPIReceiver) Start(ctx context.Context, host component.Host) error {
	address := listenOnAddress(r.port)
	r.logger.Info("Listening for requests", zap.String("address", address))

	mux := http.NewServeMux()
	mux.HandleFunc("/", r.httpHandler)
	r.httpServer = &http.Server{Addr: address, Handler: mux}
	go func() {
		_ = r.httpServer.ListenAndServe()
	}()

	telemetryClient := telemetryapi.NewClient(r.logger)
	if len(r.types) > 0 {
		_, err := telemetryClient.Subscribe(ctx, r.types, r.extensionID, fmt.Sprintf("http://%s/", address))
		if err != nil {
			r.logger.Error("Cannot register Telemetry API client", zap.Error(err))
			return err
		}
	}
	return nil
}

func (r *telemetryAPIReceiver) Shutdown(ctx context.Context) error {
	return nil
}

func newSpanID() pcommon.SpanID {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	sid := pcommon.SpanID{}
	_, _ = randSource.Read(sid[:])
	return sid
}

func newTraceID() pcommon.TraceID {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))
	tid := pcommon.TraceID{}
	_, _ = randSource.Read(tid[:])
	return tid
}

// httpHandler handles the requests coming from the Telemetry API.
// Everytime Telemetry API sends events, this function will read them from the response body
// and put into a synchronous queue to be dispatched later.
// Logging or printing besides the error cases below is not recommended if you have subscribed to
// receive extension logs. Otherwise, logging here will cause Telemetry API to send new logs for
// the printed lines which may create an infinite loop.
func (r *telemetryAPIReceiver) httpHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		r.logger.Error("error reading body", zap.Error(err))
		return
	}

	var slice []event
	if err := json.Unmarshal(body, &slice); err != nil {
		r.logger.Error("error unmarshalling body", zap.Error(err))
		return
	}

	// traces
	if r.nextTraces != nil {
		if traces, err := r.createTraces(slice); err == nil {
			if traces.SpanCount() > 0 {
				err := r.nextTraces.ConsumeTraces(context.Background(), traces)
				if err != nil {
					r.logger.Error("error receiving traces", zap.Error(err))
				}
			}
		}
	}

	// metrics
	if r.nextMetrics != nil {
		if metrics, err := r.createMetrics(slice); err == nil {
			if metrics.DataPointCount() > 0 {
				err := r.nextMetrics.ConsumeMetrics(context.Background(), metrics)
				if err != nil {
					r.logger.Error("error receiving metrics", zap.Error(err))
				}
			}
		}
	}

	// logs
	if r.nextLogs != nil {
		if logs, err := r.createLogs(slice); err == nil {
			if logs.LogRecordCount() > 0 {
				err := r.nextLogs.ConsumeLogs(context.Background(), logs)
				if err != nil {
					r.logger.Error("error receiving logs", zap.Error(err))
				}
			}
		}
	}

	r.logger.Debug("logEvents received", zap.Int("count", len(slice)), zap.Int64("queue_length", r.queue.Len()))
	slice = nil
}

func (r *telemetryAPIReceiver) createTraces(slice []event) (ptrace.Traces, error) {
	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		switch el.Type {
		// Function initialization started.
		case string(telemetryapi.PlatformInitStart):
			r.logger.Info(fmt.Sprintf("Init start: %s", r.lastPlatformStartTime), zap.Any("event", el))
			r.lastPlatformStartTime = el.Time
		// Function initialization completed.
		case string(telemetryapi.PlatformInitRuntimeDone):
			r.logger.Info(fmt.Sprintf("Init end: %s", r.lastPlatformEndTime), zap.Any("event", el))
			r.lastPlatformEndTime = el.Time
		}
		// TODO: add support for additional events, see https://docs.aws.amazon.com/lambda/latest/dg/telemetry-api.html
		// A report of function initialization.
		// case "platform.initReport":
		// Function invocation started.
		// case "platform.start":
		// The runtime finished processing an event with either success or failure.
		// case "platform.runtimeDone":
		// A report of function invocation.
		// case "platform.report":
		// Runtime restore started (reserved for future use)
		// case "platform.restoreStart":
		// Runtime restore completed (reserved for future use)
		// case "platform.restoreRuntimeDone":
		// Report of runtime restore (reserved for future use)
		// case "platform.restoreReport":
		// The extension subscribed to the Telemetry API.
		// case "platform.telemetrySubscription":
		// Lambda dropped log entries.
		// case "platform.logsDropped":
	}
	if len(r.lastPlatformStartTime) > 0 && len(r.lastPlatformEndTime) > 0 {
		td, err := r.createPlatformInitSpan(r.lastPlatformStartTime, r.lastPlatformEndTime)
		if err == nil {
			r.lastPlatformEndTime = ""
			r.lastPlatformStartTime = ""
		}
		return td, err
	}

	return ptrace.Traces{}, errors.New("no traces created")
}

func (r *telemetryAPIReceiver) createMetrics(slice []event) (pmetric.Metrics, error) {
	metrics := pmetric.NewMetrics()
	resourceMetric := metrics.ResourceMetrics().AppendEmpty()
	r.resource.CopyTo(resourceMetric.Resource())
	scopeMetric := resourceMetric.ScopeMetrics().AppendEmpty()
	scopeMetric.Scope().SetName(scopeName)
	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		if r.metricsStartTime.IsZero() {
			if t, err := time.Parse(time.RFC3339, el.Time); err == nil {
				r.metricsStartTime = t
			} else {
				return pmetric.Metrics{}, err
			}

		}
		switch el.Type {
		case string(telemetryapi.PlatformInitReport):
			jsonStr, err := json.Marshal(el.Record)
			if err != nil {
				return pmetric.Metrics{}, err
			}
			var report platformInitReport
			if err := json.Unmarshal(jsonStr, &report); err != nil {
				return pmetric.Metrics{}, err
			} else {
				if report.Phase == initPhaseInit {
					r.coldStartCounter++
					metrics := scopeMetric.Metrics().AppendEmpty()
					metrics.Metadata().PutStr("type", el.Type)
					metrics.SetName(semconv.AttributeFaaSColdstart)
					sum := metrics.SetEmptySum()
					sumHelper(sum, r.coldStartCounter, r.metricsStartTime)
				}
			}
		case string(telemetryapi.PlatformReport):
			r.invocationsCounter++
			metrics := scopeMetric.Metrics().AppendEmpty()
			metrics.Metadata().PutStr("type", el.Type)
			metrics.SetName("faas.invocations")
			sum := metrics.SetEmptySum()
			sumHelper(sum, r.invocationsCounter, r.metricsStartTime)
			jsonStr, err := json.Marshal(el.Record)
			if err != nil {
				return pmetric.Metrics{}, err
			}
			var report platformReport
			if err := json.Unmarshal(jsonStr, &report); err != nil {
				return pmetric.Metrics{}, err
			} else {
				if report.Status != statusSuccess {
					r.errorsCounter++
					metrics := scopeMetric.Metrics().AppendEmpty()
					metrics.Metadata().PutStr("type", el.Type)
					metrics.SetName("faas.errors")
					sum := metrics.SetEmptySum()
					sumHelper(sum, r.errorsCounter, r.metricsStartTime)
				}
				if report.Status == statusTimeout {
					r.timeoutsCounter++
					metrics := scopeMetric.Metrics().AppendEmpty()
					metrics.Metadata().PutStr("type", el.Type)
					metrics.SetName("faas.timeouts")
					sum := metrics.SetEmptySum()
					sumHelper(sum, r.timeoutsCounter, r.metricsStartTime)
				}
			}
		}
	}

	return metrics, nil
}

func (r *telemetryAPIReceiver) createLogs(slice []event) (plog.Logs, error) {
	logs := plog.NewLogs()
	resourceLog := logs.ResourceLogs().AppendEmpty()
	r.resource.CopyTo(resourceLog.Resource())
	scopeLog := resourceLog.ScopeLogs().AppendEmpty()
	scopeLog.Scope().SetName(scopeName)
	for _, el := range slice {
		r.logger.Debug(fmt.Sprintf("Event: %s", el.Type), zap.Any("event", el))
		logRecord := scopeLog.LogRecords().AppendEmpty()
		logRecord.Attributes().PutStr("type", el.Type)
		if t, err := time.Parse(time.RFC3339, el.Time); err == nil {
			logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
			logRecord.SetObservedTimestamp(pcommon.NewTimestampFromTime(time.Now()))
		} else {
			r.logger.Error("error parsing time", zap.Error(err))
			return plog.Logs{}, err
		}
		if el.Type == string(telemetryapi.Function) || el.Type == string(telemetryapi.Extension) {
			if record, ok := el.Record.(map[string]interface{}); ok {
				// in JSON format https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
				if timestamp, ok := record["timestamp"].(string); ok {
					if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
						logRecord.SetTimestamp(pcommon.NewTimestampFromTime(t))
					} else {
						// Just print a debug message
						r.logger.Debug("error parsing time", zap.Error(err))
					}
				}
				if level, ok := record["level"].(string); ok {
					logRecord.SetSeverityNumber(severityTextToNumber(strings.ToUpper(level)))
					logRecord.SetSeverityText(logRecord.SeverityNumber().String())
				}
				if requestId, ok := record["requestId"].(string); ok {
					logRecord.Attributes().PutStr(semconv.AttributeFaaSInvocationID, requestId)
				}
				if line, ok := record["message"].(string); ok {
					logRecord.Body().SetStr(line)
				}
			} else {
				// in plain text https://docs.aws.amazon.com/lambda/latest/dg/telemetry-schema-reference.html#telemetry-api-function
				if line, ok := el.Record.(string); ok {
					logRecord.Body().SetStr(line)
				}
			}
		}
	}
	return logs, nil
}

func sumHelper(sum pmetric.Sum, count int64, metricsStartTime time.Time) {
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)
	dp := sum.DataPoints().AppendEmpty()
	dp.SetIntValue(count)
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(metricsStartTime))
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.Attributes().PutStr(semconv.AttributeFaaSTrigger, semconv.AttributeFaaSTriggerOther)
}

func severityTextToNumber(severityText string) plog.SeverityNumber {
	mapping := map[string]plog.SeverityNumber{
		"TRACE":    plog.SeverityNumberTrace,
		"TRACE2":   plog.SeverityNumberTrace2,
		"TRACE3":   plog.SeverityNumberTrace3,
		"TRACE4":   plog.SeverityNumberTrace4,
		"DEBUG":    plog.SeverityNumberDebug,
		"DEBUG2":   plog.SeverityNumberDebug2,
		"DEBUG3":   plog.SeverityNumberDebug3,
		"DEBUG4":   plog.SeverityNumberDebug4,
		"INFO":     plog.SeverityNumberInfo,
		"INFO2":    plog.SeverityNumberInfo2,
		"INFO3":    plog.SeverityNumberInfo3,
		"INFO4":    plog.SeverityNumberInfo4,
		"WARN":     plog.SeverityNumberWarn,
		"WARN2":    plog.SeverityNumberWarn2,
		"WARN3":    plog.SeverityNumberWarn3,
		"WARN4":    plog.SeverityNumberWarn4,
		"ERROR":    plog.SeverityNumberError,
		"ERROR2":   plog.SeverityNumberError2,
		"ERROR3":   plog.SeverityNumberError3,
		"ERROR4":   plog.SeverityNumberError4,
		"FATAL":    plog.SeverityNumberFatal,
		"FATAL2":   plog.SeverityNumberFatal2,
		"FATAL3":   plog.SeverityNumberFatal3,
		"FATAL4":   plog.SeverityNumberFatal4,
		"CRITICAL": plog.SeverityNumberFatal,
		"ALL":      plog.SeverityNumberTrace,
		"WARNING":  plog.SeverityNumberWarn,
	}
	if ans, ok := mapping[strings.ToUpper(severityText)]; ok {
		return ans
	} else {
		return plog.SeverityNumberUnspecified
	}
}

func (r *telemetryAPIReceiver) registerTracesConsumer(next consumer.Traces) {
	r.nextTraces = next
}

func (r *telemetryAPIReceiver) registerMetricsConsumer(next consumer.Metrics) {
	r.nextMetrics = next
}

func (r *telemetryAPIReceiver) registerLogsConsumer(next consumer.Logs) {
	r.nextLogs = next
}

func (r *telemetryAPIReceiver) createPlatformInitSpan(start, end string) (ptrace.Traces, error) {
	traceData := ptrace.NewTraces()
	rs := traceData.ResourceSpans().AppendEmpty()
	r.resource.CopyTo(rs.Resource())

	ss := rs.ScopeSpans().AppendEmpty()
	ss.Scope().SetName(scopeName)
	span := ss.Spans().AppendEmpty()
	span.SetTraceID(newTraceID())
	span.SetSpanID(newSpanID())
	span.SetName("platform.initRuntimeDone")
	span.SetKind(ptrace.SpanKindInternal)
	span.Attributes().PutBool(semconv.AttributeFaaSColdstart, true)
	startTime, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
	endTime, err := time.Parse(time.RFC3339, end)
	if err != nil {
		return ptrace.Traces{}, err
	}
	span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))
	return traceData, nil
}

func newTelemetryAPIReceiver(
	cfg *Config,
	set receiver.Settings,
) (*telemetryAPIReceiver, error) {
	envResourceMap := map[string]string{
		"AWS_LAMBDA_FUNCTION_VERSION": semconv.AttributeFaaSVersion,
		"AWS_REGION":                  semconv.AttributeFaaSInvokedRegion,
	}
	r := pcommon.NewResource()
	r.Attributes().PutStr(semconv.AttributeFaaSInvokedProvider, semconv.AttributeFaaSInvokedProviderAWS)
	if val, ok := os.LookupEnv("OTEL_SERVICE_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
	} else if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeServiceName, val)
	} else {
		r.Attributes().PutStr(semconv.AttributeServiceName, "unknown_service")
	}
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		r.Attributes().PutStr(semconv.AttributeFaaSName, val)
	}
	if val, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_MEMORY_SIZE"); ok {
		if mb, err := strconv.Atoi(val); err == nil {
			r.Attributes().PutInt(semconv.AttributeFaaSMaxMemory, int64(mb)*1024*1024)
		}
	}
	// https://opentelemetry.io/docs/specs/otel/metrics/data-model/#single-writer
	r.Attributes().PutStr(semconv.AttributeServiceInstanceID, uuid.New().String())

	for env, resourceAttribute := range envResourceMap {
		if val, ok := os.LookupEnv(env); ok {
			r.Attributes().PutStr(resourceAttribute, val)
		}
	}

	subscribedTypes := []telemetryapi.EventType{}
	for _, val := range cfg.Types {
		switch val {
		case "platform":
			subscribedTypes = append(subscribedTypes, telemetryapi.Platform)
		case "function":
			subscribedTypes = append(subscribedTypes, telemetryapi.Function)
		case "extension":
			subscribedTypes = append(subscribedTypes, telemetryapi.Extension)
		}
	}

	return &telemetryAPIReceiver{
		logger:             set.Logger,
		queue:              queue.New(initialQueueSize),
		extensionID:        cfg.extensionID,
		port:               cfg.Port,
		types:              subscribedTypes,
		resource:           r,
		coldStartCounter:   0,
		errorsCounter:      0,
		invocationsCounter: 0,
		timeoutsCounter:    0,
	}, nil
}

func listenOnAddress(port int) string {
	envAwsLocal, ok := os.LookupEnv("AWS_SAM_LOCAL")
	var addr string
	if ok && envAwsLocal == "true" {
		addr = ":" + strconv.Itoa(port)
	} else {
		addr = "sandbox.localdomain:" + strconv.Itoa(port)
	}

	return addr
}
