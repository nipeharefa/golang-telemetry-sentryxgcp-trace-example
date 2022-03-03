package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// HTTP middleware setting a value on the request context
func MyMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		opts := []trace.SpanStartOption{
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...)}

		ctx, span := tracer.Start(r.Context(), r.URL.String(), opts...)
		// t1 := time.Now()
		defer func() {
			defer span.End()
			fmt.Println("test")
		}()

		rr := r.WithContext(ctx)

		next.ServeHTTP(w, rr)
	}

	return http.HandlerFunc(fn)
}

func Resource() *resource.Resource {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		// semconv.ServiceNameKey.String("stdout-example"),
		// semconv.ServiceVersionKey.String("0.0.1"),
		semconv.ServiceNameKey.String("go-metrics"),
	)
}

func CreateStdoutTrace() *stdouttrace.Exporter {

	exporter1, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdout exporter: %v", err)
	}

	return exporter1
}

func CreateOTLPNewRelic(ctx context.Context) *otlptrace.Exporter {
	client := otlptracehttp.NewClient(otlptracehttp.WithHeaders(map[string]string{
		"api-key": os.Getenv("SENTRY_API_KEY"),
	}))
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatalf("creating OTLP trace exporter: %v", err)
	}

	return exporter
}

func CreateGoogleCloudTraceEXporter() *texporter.Exporter {

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	exporter, err := texporter.New(texporter.WithProjectID(projectID))
	if err != nil {
		log.Fatalf("texporter.NewExporter: %v", err)
	}
	return exporter
}

func main() {
	ctx := context.Background()
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM, os.Kill)

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(CreateOTLPNewRelic(ctx)), sdktrace.WithResource(Resource()))
	defer tp.ForceFlush(ctx) // flushes any pending spans
	otel.SetTracerProvider(tp)
	tracer = otel.GetTracerProvider().Tracer("")

	r := chi.NewRouter()
	r.Use(MyMiddleware)
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {

		time.Sleep(100 * time.Millisecond)
		_, span := tracer.Start(r.Context(), "child")
		defer span.End()

		span.AddEvent("Acquiring lock", trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", r)...))
		time.Sleep(200 * time.Millisecond)
		span.AddEvent("Unlocking")
		w.Write([]byte("welcome"))
	})

	srv := &http.Server{
		Addr:    "0.0.0.0:3000",
		Handler: r,
	}

	go func() {
		fmt.Println("Starting server on port 3000")
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	<-stopSignal

	fmt.Println("Shutting down the server... ⌛️")
	err := srv.Shutdown(context.Background())
	if err != nil {
		// log.Fatal()
		fmt.Println(err)
	}
	err = tp.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Shutting down the server...Done ✅")
}
