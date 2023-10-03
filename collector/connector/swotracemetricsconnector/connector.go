package swotracemetricsconnector

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"time"
)

type connectorImp struct {
	config          Config
	metricsConsumer consumer.Metrics
	logger          *zap.Logger
	component.StartFunc
	component.ShutdownFunc
}

func newConnector(logger *zap.Logger, config component.Config) (*connectorImp, error) {
	logger.Info("Building swo-trace-metrics-connector")
	cfg := config.(*Config)
	return &connectorImp{
		config: *cfg,
		logger: logger,
	}, nil
}

func (con *connectorImp) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func (con *connectorImp) ConsumeTraces(ctx context.Context, traceData ptrace.Traces) error {
	for i := 0; i < traceData.ResourceSpans().Len(); i++ {
		resourceSpan := traceData.ResourceSpans().At(i)

		for j := 0; j < resourceSpan.ScopeSpans().Len(); j++ {
			scopeSpan := resourceSpan.ScopeSpans().At(j)

			for k := 0; k < scopeSpan.Spans().Len(); k++ {
				span := scopeSpan.Spans().At(k)

				if span.ParentSpanID().IsEmpty() {
					startTime := span.StartTimestamp()
					endTime := span.EndTimestamp()

					if endTime > startTime {
						duration := uint64(endTime-startTime) / 1000
						metrics := createMetric(&resourceSpan, &span, duration)
						if err := con.metricsConsumer.ConsumeMetrics(ctx, metrics); err != nil {
							con.logger.Error("Failed ConsumeMetrics", zap.Error(err))
						}
					}
				}
			}
		}
	}
	return nil
}

func createMetric(resourceSpan *ptrace.ResourceSpans, span *ptrace.Span, duration uint64) pmetric.Metrics {
	metrics := pmetric.NewMetrics()
	resourceMetric := metrics.ResourceMetrics().AppendEmpty()

	spanAttrs := span.Attributes()
	resourceSpan.Resource().Attributes().CopyTo(resourceMetric.Resource().Attributes())
	resourceMetric.Resource().Attributes().PutStr("sw.trace_span_mode", "otel")

	scopeMetric := resourceMetric.ScopeMetrics().AppendEmpty()
	scopeMetric.Scope().SetName("swotracemetricsconnector")
	responseTimeMetric := scopeMetric.Metrics().AppendEmpty()

	responseTimeMetric.SetName("ResponseTime")
	responseTimeMetric.SetEmptySum().SetIsMonotonic(false)
	responseTimeMetric.Sum().SetAggregationTemporality(pmetric.AggregationTemporalityDelta)

	dps := responseTimeMetric.Sum().DataPoints()
	dps.EnsureCapacity(1)
	timestamp := pcommon.NewTimestampFromTime(time.Now())

	dp := dps.AppendEmpty()
	dp.SetStartTimestamp(timestamp)
	dp.SetTimestamp(timestamp)

	dpAttrs := dp.Attributes()
	status := span.Status()
	hasError := status.Code() == ptrace.StatusCodeError

	dp.SetIntValue(int64(duration))
	addSwoDimensions(&dpAttrs, &spanAttrs, hasError)
	return metrics
}

func addSwoDimensions(attributes *pcommon.Map, spanAttr *pcommon.Map, hasError bool) {
	httpStatus, ok := spanAttr.Get(string(semconv.HTTPStatusCodeKey))
	if !hasError && ok {
		code := httpStatus.Int()
		hasError = int(code/100) == 5
	}

	if ok {
		attributes.PutStr("http.status_code", httpStatus.Str())
	}

	if httpMethod, exist := spanAttr.Get(string(semconv.HTTPMethodKey)); exist {
		attributes.PutStr("http.method", httpMethod.Str())
	}

	if hasError {
		attributes.PutStr("sw.is_error", "true")
	} else {
		attributes.PutStr("sw.is_error", "false")
	}
}
