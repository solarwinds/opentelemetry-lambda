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

type event struct {
	Time   string `json:"time"`
	Type   string `json:"type"`
	Record any    `json:"record"`
}

type initPhase string

const (
	initPhaseInit   initPhase = "init"
	initPhaseInvoke initPhase = "invoke"
)

type initReportMetrics struct {
	DurationMs float64 `json:"durationMs"`
}

type initType string

const (
	initTypeOnDemand               initType = "on-demand"
	initTypeProvisionedConcurrency initType = "provisioned-concurrency"
	initTypeSnapStart              initType = "snap-start"
)

type status string

const (
	statusSuccess status = "success"
	statusError   status = "error"
	statusFailure status = "failure"
	statusTimeout status = "timeout"
)

type platformExtension struct {
	Events    []string `json:"events"`
	ErrorType *string  `json:"errorType"`
	Name      string   `json:"name"`
	State     string   `json:"state"`
}

type platformInitReport struct {
	Status             status            `json:"status"`
	ErrorType          string            `json:"errorType"`
	InitializationType initType          `json:"initializationType"`
	Metrics            initReportMetrics `json:"metrics"`
	Phase              initPhase         `json:"phase"`
	Spans              []span            `json:"spans"`
	Tracing            *traceContext     `json:"tracing"`
}

type platformInitRuntimeDone struct {
	Status             status        `json:"status"`
	ErrorType          string        `json:"errorType"`
	InitializationType initType      `json:"initializationType"`
	Phase              initPhase     `json:"phase"`
	Spans              []span        `json:"spans"`
	Tracing            *traceContext `json:"tracing"`
}

type platformInitStart struct {
	InitializationType initType      `json:"initializationType"`
	Phase              initPhase     `json:"phase"`
	FunctionName       string        `json:"functionName"`
	FunctionVersion    string        `json:"functionVersion"`
	InstanceId         string        `json:"instanceId"`
	InstanceMaxMemory  uint32        `json:"instanceMaxMemory"`
	RuntimeVersion     string        `json:"runtimeVersion"`
	RuntimeVersionArn  string        `json:"runtimeVersionArn"`
	Tracing            *traceContext `json:"tracing"`
}

type platformLogsDropped struct {
	DroppedBytes   uint64 `json:"droppedBytes"`
	DroppedRecords uint64 `json:"droppedRecords"`
	Reason         string `json:"reason"`
}

type platformReport struct {
	Status    status        `json:"status"`
	ErrorType string        `json:"errorType"`
	Metrics   reportMetrics `json:"metrics"`
	RequestId string        `json:"requestId"`
	Spans     []span        `json:"spans"`
	Tracing   *traceContext `json:"tracing"`
}

type platformRestoreReport struct {
	Status    status               `json:"status"`
	ErrorType string               `json:"errorType"`
	Metrics   restoreReportMetrics `json:"metrics"`
	Spans     []span               `json:"spans"`
	Tracing   *traceContext        `json:"tracing"`
}

type platformRestoreRuntimeDone struct {
	Status    status        `json:"status"`
	ErrorType string        `json:"errorType"`
	Spans     []span        `json:"spans"`
	Tracing   *traceContext `json:"tracing"`
}

type platformRestoreStart struct {
	FunctionName      string        `json:"functionName"`
	FunctionVersion   string        `json:"functionVersion"`
	InstanceId        string        `json:"instanceId"`
	InstanceMaxMemory uint32        `json:"instanceMaxMemory"`
	RuntimeVersion    string        `json:"runtimeVersion"`
	RuntimeVersionArn string        `json:"runtimeVersionArn"`
	Tracing           *traceContext `json:"tracing"`
}

type platformRuntimeDone struct {
	Status    status              `json:"status"`
	ErrorType string              `json:"errorType"`
	RequestId string              `json:"requestId"`
	Metrics   *runtimeDoneMetrics `json:"metrics"`
	Spans     []span              `json:"spans"`
	Tracing   *traceContext       `json:"tracing"`
}

type platformStart struct {
	RequestId string        `json:"requestId"`
	Tracing   *traceContext `json:"tracing"`
	Version   *string       `json:"version"`
}

type platformTelemetrySubscription struct {
	Name  string                  `json:"name"`
	State string                  `json:"state"`
	Types []subscriptionEventType `json:"types"`
}

type reportMetrics struct {
	BilledDurationMs        uint64   `json:"billedDurationMs"`
	BilledRestoreDurationMs *uint64  `json:"billedRestoreDurationMs"`
	DurationMs              float64  `json:"durationMs"`
	InitDurationMs          *float64 `json:"initDurationMs"`
	MaxMemoryUsedMB         uint64   `json:"maxMemoryUsedMB"`
	MemorySizeMB            uint64   `json:"memorySizeMB"`
	RestoreDurationMs       *float64 `json:"restoreDurationMs"`
}

type restoreReportMetrics struct {
	DurationMs float64 `json:"durationMs"`
}

type runtimeDoneMetrics struct {
	DurationMs    float64 `json:"durationMs"`
	ProducedBytes *uint64 `json:"producedBytes"`
}

type span struct {
	DurationMs float64 `json:"durationMs"`
	Name       string  `json:"name"`
	Start      string  `json:"start"`
}

type subscriptionEventType string

const (
	subscriptionEventTypePlatform  subscriptionEventType = "platform"
	subscriptionEventTypeFunction  subscriptionEventType = "function"
	subscriptionEventTypeExtension subscriptionEventType = "extension"
)

type traceContext struct {
	SpanID *string     `json:"spanId"`
	Type   tracingType `json:"type"`
	Value  string      `json:"value"`
}

type tracingType string

const (
	tracingTypeXAmznTraceId tracingType = "X-Amzn-Trace-Id"
)
