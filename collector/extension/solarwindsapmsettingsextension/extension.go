package solarwindsapmsettingsextension

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"github.com/solarwindscloud/apm-proto/go/collectorpb"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
	"math"
	"os"
	"strconv"
	"time"
)

const (
	JSONOutputFile      = "/tmp/solarwinds-apm-settings.json"
	GrpcContextDeadline = time.Duration(1) * time.Second
	Cert                = `-----BEGIN CERTIFICATE-----
MIIGeDCCBWCgAwIBAgIQA3Oe8N8Yjiru3H/RGzOVkDANBgkqhkiG9w0BAQsFADA8
MQswCQYDVQQGEwJVUzEPMA0GA1UEChMGQW1hem9uMRwwGgYDVQQDExNBbWF6b24g
UlNBIDIwNDggTTAzMB4XDTI0MDEyMTAwMDAwMFoXDTI1MDIxOTIzNTk1OVowJzEl
MCMGA1UEAwwcKi5uYS0wMS5jbG91ZC5zb2xhcndpbmRzLmNvbTCCASIwDQYJKoZI
hvcNAQEBBQADggEPADCCAQoCggEBAJzFsCYa1dmTHgjIpY0FMdsatA8+ipChW/FD
R2fRSjVx4SAOl+RNhuXnWIpFwLOp6kr9/T6XEagu4NHoPqdRjx7wsij201FuciJd
GbHQO17a2amklHCrAoyyTd3sld5U8RT4vRUtWTFR0zX+fLJrCS/Ecb0O92xP8J0p
Rts9D5SBWHgFk5NkB6INl+EVVmwhMUHHdhHtimmuyISuCKdLCdwrUrHoLj6LI1aG
WvNLhPxBK5QMLqGBXwzmjJPNM5+fIr833WGf7San3+DtF+ACN+kvWYBwEOoWVEsb
C9upzkgQ0mrqij50bu4xlznAQv8EyxZvhVUoCLnBRpq44eAcCw8CAwEAAaOCA4kw
ggOFMB8GA1UdIwQYMBaAFFXZGF/SHMwB4Vi0vqvZVUIB1y4CMB0GA1UdDgQWBBQz
puQomVZ2Ft+6WcSyWLVHO4EZfzCBuQYDVR0RBIGxMIGughwqLm5hLTAxLmNsb3Vk
LnNvbGFyd2luZHMuY29tgiYqLmNvbGxlY3Rvci5kYy0wMS5jbG91ZC5zb2xhcndp
bmRzLmNvbYIgKi5jb2xsZWN0b3IuY2xvdWQuc29sYXJ3aW5kcy5jb22CHCouZGMt
MDEuY2xvdWQuc29sYXJ3aW5kcy5jb22CJiouY29sbGVjdG9yLm5hLTAxLmNsb3Vk
LnNvbGFyd2luZHMuY29tMBMGA1UdIAQMMAowCAYGZ4EMAQIBMA4GA1UdDwEB/wQE
AwIFoDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwOwYDVR0fBDQwMjAw
oC6gLIYqaHR0cDovL2NybC5yMm0wMy5hbWF6b250cnVzdC5jb20vcjJtMDMuY3Js
MHUGCCsGAQUFBwEBBGkwZzAtBggrBgEFBQcwAYYhaHR0cDovL29jc3AucjJtMDMu
YW1hem9udHJ1c3QuY29tMDYGCCsGAQUFBzAChipodHRwOi8vY3J0LnIybTAzLmFt
YXpvbnRydXN0LmNvbS9yMm0wMy5jZXIwDAYDVR0TAQH/BAIwADCCAX8GCisGAQQB
1nkCBAIEggFvBIIBawFpAHcATnWjJ1yaEMM4W2zU3z9S6x3w4I4bjWnAsfpksWKa
Od8AAAGNKVzSFQAABAMASDBGAiEA9vyqHQW6/weVAF+M0fF2nAFWny8zOxnlKBZh
UIGav6wCIQDj5pdnkH20v5x3yeRI3U4G+Gd5tzUMSKC9DgegE402HQB2AH1ZHhLh
eCp7HGFnfF79+NCHXBSgTpWeuQMv2Q6MLnm4AAABjSlc0kAAAAQDAEcwRQIgc1HX
ZloOkamcWsgICGJZMGroKtdVSr1ICAD9m+EmJWsCIQCeran8qAK8nH6uHtBK8mjs
tEIf8iR8SMAhgmS2Jaw8QQB2AObSMWNAd4zBEEEG13G5zsHSQPaWhIb7uocyHf0e
N45QAAABjSlc0l4AAAQDAEcwRQIhAK4CV77Lwm3PNQLw2PaSPEFmJIKVoefHyGxF
FJ7FPWe2AiAMbJAlGLX8e7+Mrl5AvWnFzFLzA8sZLHkJBHirQ/XX0jANBgkqhkiG
9w0BAQsFAAOCAQEAitigIgKeHfs/AzjcrsBp/jyVdCAykSMHfkMlqklArSfzfXtt
Ljo1Kfri76Mf5VSMWiRc1toaVNlNwn6+lWL276pPgzKq8RCemGu5gsrlHqLsJ1kX
je+ou/+2NPpBURd868jmLiiNscCcp4UHWPpfaoeijIGb5G1BStiSzvaYe4OGTkTA
AvJ87WeatZOmgknRDIRKqz7jIlMQgSI8iycxS+/kBOUO7KzsJp30jQTXi8UK2riM
j33Li5/9wBKv8q0cy5uNwHokP2IuIC6OkO0wF0Q69L1pOttftFbUMn6xexjJbUru
7FZXpJ1569+Yk28hNbeIqqu+cia0zgpz26pOkg==
-----END CERTIFICATE-----`
)

type solarwindsapmSettingsExtension struct {
	logger *zap.Logger
	config *Config
	cancel context.CancelFunc
	conn   *grpc.ClientConn
	client collectorpb.TraceCollectorClient
}

func newSolarwindsApmSettingsExtension(extensionCfg *Config, logger *zap.Logger) (extension.Extension, error) {
	settingsExtension := &solarwindsapmSettingsExtension{
		config: extensionCfg,
		logger: logger,
	}
	return settingsExtension, nil
}

func refresh(extension *solarwindsapmSettingsExtension) {
	extension.logger.Info("Time to refresh from " + extension.config.Endpoint)
	if hostname, err := os.Hostname(); err != nil {
		extension.logger.Error("Unable to call os.Hostname() " + err.Error())
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), GrpcContextDeadline)
		defer cancel()

		request := &collectorpb.SettingsRequest{
			ApiKey: extension.config.Key,
			Identity: &collectorpb.HostID{
				Hostname: hostname,
			},
			ClientVersion: "2",
		}
		if response, err := extension.client.GetSettings(ctx, request); err != nil {
			extension.logger.Error("Unable to getSettings from " + extension.config.Endpoint + " " + err.Error())
		} else {
			switch result := response.GetResult(); result {
			case collectorpb.ResultCode_OK:
				if len(response.GetWarning()) > 0 {
					extension.logger.Warn(response.GetWarning())
				}
				var settings []map[string]interface{}
				for _, item := range response.GetSettings() {
					marshalOptions := protojson.MarshalOptions{
						UseEnumNumbers:  true,
						EmitUnpopulated: true,
					}
					if settingBytes, err := marshalOptions.Marshal(item); err != nil {
						extension.logger.Warn("Error to marshal setting JSON[] byte from response.GetSettings() " + err.Error())
					} else {
						setting := make(map[string]interface{})
						if err := json.Unmarshal(settingBytes, &setting); err != nil {
							extension.logger.Warn("Error to unmarshal setting JSON object from setting JSON[]byte " + err.Error())
						} else {
							if value, ok := setting["value"].(string); ok {
								if num, e := strconv.ParseInt(value, 10, 0); e != nil {
									extension.logger.Warn("Unable to parse value " + value + " as number " + e.Error())
								} else {
									setting["value"] = num
								}
							}
							if timestamp, ok := setting["timestamp"].(string); ok {
								if num, e := strconv.ParseInt(timestamp, 10, 0); e != nil {
									extension.logger.Warn("Unable to parse timestamp " + timestamp + " as number " + e.Error())
								} else {
									setting["timestamp"] = num
								}
							}
							if ttl, ok := setting["ttl"].(string); ok {
								if num, e := strconv.ParseInt(ttl, 10, 0); e != nil {
									extension.logger.Warn("Unable to parse ttl " + ttl + " as number " + e.Error())
								} else {
									setting["ttl"] = num
								}
							}
							if _, ok := setting["flags"]; ok {
								setting["flags"] = string(item.Flags)
							}
							if arguments, ok := setting["arguments"].(map[string]interface{}); ok {
								if value, ok := item.Arguments["BucketCapacity"]; ok {
									arguments["BucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["BucketRate"]; ok {
									arguments["BucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["TriggerRelaxedBucketCapacity"]; ok {
									arguments["TriggerRelaxedBucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["TriggerRelaxedBucketRate"]; ok {
									arguments["TriggerRelaxedBucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["TriggerStrictBucketCapacity"]; ok {
									arguments["TriggerStrictBucketCapacity"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["TriggerStrictBucketRate"]; ok {
									arguments["TriggerStrictBucketRate"] = math.Float64frombits(binary.LittleEndian.Uint64(value))
								}
								if value, ok := item.Arguments["MetricsFlushInterval"]; ok {
									arguments["MetricsFlushInterval"] = int32(binary.LittleEndian.Uint32(value))
								}
								if value, ok := item.Arguments["MaxTransactions"]; ok {
									arguments["MaxTransactions"] = int32(binary.LittleEndian.Uint32(value))
								}
								if value, ok := item.Arguments["MaxCustomMetrics"]; ok {
									arguments["MaxCustomMetrics"] = int32(binary.LittleEndian.Uint32(value))
								}
								if value, ok := item.Arguments["EventsFlushInterval"]; ok {
									arguments["EventsFlushInterval"] = int32(binary.LittleEndian.Uint32(value))
								}
								if value, ok := item.Arguments["ProfilingInterval"]; ok {
									arguments["ProfilingInterval"] = int32(binary.LittleEndian.Uint32(value))
								}
								// Remove SignatureKey from collector response
								delete(arguments, "SignatureKey")
							}
							settings = append(settings, setting)
						}
					}
				}
				if content, err := json.Marshal(settings); err != nil {
					extension.logger.Warn("Error to marshal setting JSON[] byte from settings " + err.Error())
				} else {
					if err := os.WriteFile(JSONOutputFile, content, 0644); err != nil {
						extension.logger.Error("Unable to write " + JSONOutputFile + " " + err.Error())
					} else {
						if len(response.GetWarning()) > 0 {
							extension.logger.Warn(JSONOutputFile + " is refreshed (soft disabled)")
						} else {
							extension.logger.Info(JSONOutputFile + " is refreshed")
						}
						extension.logger.Info(string(content))
					}
				}
			case collectorpb.ResultCode_TRY_LATER:
				extension.logger.Warn("GetSettings returned TRY_LATER " + response.GetWarning())
			case collectorpb.ResultCode_INVALID_API_KEY:
				extension.logger.Warn("GetSettings returned INVALID_API_KEY " + response.GetWarning())
			case collectorpb.ResultCode_LIMIT_EXCEEDED:
				extension.logger.Warn("GetSettings returned LIMIT_EXCEEDED " + response.GetWarning())
			case collectorpb.ResultCode_REDIRECT:
				extension.logger.Warn("GetSettings returned REDIRECT " + response.GetWarning())
			default:
				extension.logger.Warn("Unknown ResultCode from GetSettings " + response.GetWarning())
			}
		}
	}
}

func (extension *solarwindsapmSettingsExtension) Start(ctx context.Context, _ component.Host) error {
	extension.logger.Info("Starting up solarwinds apm settings extension")
	ctx = context.Background()
	ctx, extension.cancel = context.WithCancel(ctx)
	certPool := x509.NewCertPool()
	cert := []byte(Cert)
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		extension.logger.Error("Unable to bundle cert")
	} else {
		extension.logger.Error("Bundled the hardcoded cert")
	}
	var err error
	extension.conn, err = grpc.Dial(extension.config.Endpoint, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: certPool})))
	if err != nil {
		return err
	}
	extension.logger.Info("grpc.Dial to " + extension.config.Endpoint)
	extension.client = collectorpb.NewTraceCollectorClient(extension.conn)

	// initial refresh
	refresh(extension)

	go func() {
		ticker := time.NewTicker(extension.config.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				refresh(extension)
			case <-ctx.Done():
				extension.logger.Info("Received ctx.Done() from ticker")
				return
			}
		}
	}()

	return nil
}

func (extension *solarwindsapmSettingsExtension) Shutdown(_ context.Context) error {
	extension.logger.Info("Shutting down solarwinds apm settings extension")
	if extension.cancel != nil {
		extension.cancel()
	}
	if extension.conn != nil {
		return extension.conn.Close()
	}
	return nil
}
