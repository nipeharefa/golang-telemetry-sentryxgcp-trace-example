module github.com/nipeharefa/golang-cloud-trace-gcp-example

go 1.16

require (
	github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace v1.3.0
	github.com/go-chi/chi v1.5.4
	go.opentelemetry.io/otel v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.4.1
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.4.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.4.1
	go.opentelemetry.io/otel/sdk v1.4.1
	go.opentelemetry.io/otel/trace v1.4.1
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
)
