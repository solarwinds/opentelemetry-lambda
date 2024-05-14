package telemetryapireceiver // import "github.com/open-telemetry/opentelemetry-lambda/collector/receiver/telemetryapireceiver"

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListenOnLogsAddress(t *testing.T) {
	testCases := []struct {
		desc     string
		testFunc func(*testing.T)
	}{
		{
			desc: "listen on address without AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				addr := listenOnLogsAddress()
				require.EqualValues(t, "sandbox.localdomain:4327", addr)
			},
		},
		{
			desc: "listen on address with AWS_SAM_LOCAL env variable",
			testFunc: func(t *testing.T) {
				t.Setenv("AWS_SAM_LOCAL", "true")
				addr := listenOnLogsAddress()
				require.EqualValues(t, ":4327", addr)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, tc.testFunc)
	}
}

//type mockConsumer struct {
//	consumed int
//}

//func (c *mockConsumer) ConsumeLogs(ctx context.Context, td plog.Logs) error {
//	return nil
//}

//	func (c *mockConsumer) Capabilities() consumer.Capabilities {
//		return consumer.Capabilities{MutatesData: true}
//	}
//func TestHandler(t *testing.T) {
//	testCases := []struct {
//		desc          string
//		body          string
//		expectedSpans int
//	}{
//		{
//			desc: "empty body",
//			body: `{}`,
//		},
//		{
//			desc: "invalid json",
//			body: `invalid json`,
//		},
//		{
//			desc: "valid event",
//			body: `[{"time":"", "type":"", "record": {}}]`,
//		},
//		{
//			desc: "valid event",
//			body: `[{"time":"", "type":"platform.initStart", "record": {}}]`,
//		},
//		{
//			desc: "valid start/end events",
//			body: `[
//				{"time":"2006-01-02T15:04:04.000Z", "type":"platform.initStart", "record": {}},
//				{"time":"2006-01-02T15:04:05.000Z", "type":"platform.initRuntimeDone", "record": {}}
//			]`,
//			expectedLogs: 1,
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.desc, func(t *testing.T) {
//			consumer := mockConsumer{}
//			r, err := newTelemetryAPIReceiver(
//				&Config{},
//				&consumer,
//				receivertest.NewNopCreateSettings(),
//			)
//			require.NoError(t, err)
//			req := httptest.NewRequest("POST",
//				"http://localhost:53612/someevent", strings.NewReader(tc.body))
//			rec := httptest.NewRecorder()
//			r.httpHandler(rec, req)
//			require.Equal(t, tc.expectedSpans, consumer.consumed)
//		})
//	}
//}

//
//func TestCreatePlatformInitSpan(t *testing.T) {
//	testCases := []struct {
//		desc        string
//		start       string
//		end         string
//		expected    int
//		expectError bool
//	}{
//		{
//			desc:        "no start/end times",
//			expectError: true,
//		},
//		{
//			desc:        "no end time",
//			start:       "2006-01-02T15:04:05.000Z",
//			expectError: true,
//		},
//		{
//			desc:        "no start times",
//			end:         "2006-01-02T15:04:05.000Z",
//			expectError: true,
//		},
//		{
//			desc:        "valid times",
//			start:       "2006-01-02T15:04:04.000Z",
//			end:         "2006-01-02T15:04:05.000Z",
//			expected:    1,
//			expectError: false,
//		},
//	}
//	for _, tc := range testCases {
//		t.Run(tc.desc, func(t *testing.T) {
//			r, err := newTelemetryAPIReceiver(
//				&Config{},
//				nil,
//				receivertest.NewNopCreateSettings(),
//			)
//			require.NoError(t, err)
//			td, err := r.createPlatformInitSpan(tc.start, tc.end)
//			if tc.expectError {
//				require.Error(t, err)
//			} else {
//				require.Equal(t, tc.expected, td.SpanCount())
//			}
//		})
//	}
//}
