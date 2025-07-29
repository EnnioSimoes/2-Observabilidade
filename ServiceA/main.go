package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/EnnioSimoes/2-Observabilidade/ServiceA/configs"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cepRequest struct {
	Cep string `json:"cep"`
}

var serviceName = semconv.ServiceNameKey.String("service-a")

// Initialize a gRPC connection to be used by both the tracer and meter
// providers.
func initConn() (*grpc.ClientConn, error) {
	// It connects the OpenTelemetry Collector through local gRPC connection.
	// You may replace `localhost:4317` with your endpoint.
	conn, err := grpc.NewClient("otel-collector:4317",
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	return conn, err
}

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracerProvider(ctx context.Context, res *resource.Resource, conn *grpc.ClientConn) (func(context.Context) error, error) {
	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
	/**
	 * Instrumenta o handler com OpenTelemetry.
	 */
	// Pega o tracer global que configuramos
	tracer := otel.Tracer("service-a")

	// Inicia um novo span. O contexto da requisição é usado como pai.
	ctx, span := tracer.Start(ctx, "StartHandlerSpan")
	defer span.End() // É crucial finalizar o span

	time.Sleep(1 * time.Second) // Simula algum processamento
	/**
	 * Fim da instrumentação do handler.
	 */

	// Adiciona atributos ao span para dar mais detalhes
	span.SetAttributes(
		attribute.String("http.method", r.Method),
		attribute.String("http.url", r.URL.String()),
	)

	log.Println("Received request for:", r.URL.Path)
	var req cepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	cep := req.Cep

	if cep == "" {
		log.Println("No CEP provided in the request")
		http.Error(w, "CEP is required", http.StatusBadRequest)
		return
	}
	log.Println("Extracted CEP:", cep)

	_, err := checkCep(cep)
	if err != nil {
		log.Printf("Invalid CEP: %v\n", err)
		http.Error(w, "Invalid zipcode", http.StatusUnprocessableEntity)
		return
	}

	temperature, err := getTemperature(cep, ctx)
	if err != nil {
		log.Printf("Error getting temperature: %v\n", err)
		http.Error(w, "Failed to get temperature", http.StatusInternalServerError)
		return
	}

	log.Println("Temperature data received:", temperature)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(temperature))
	if err != nil {
		log.Printf("Error writing response: %v\n", err)
		return
	}
	log.Println("Response sent successfully")
}

func getTemperature(cep string, ctx context.Context) (string, error) {
	// Intrumenta o span para a chamada interna
	// Pega o tracer novamente (ou poderia ser passado como argumento)
	tracer := otel.Tracer("service-a")

	// Inicia um span filho, pois estamos usando o contexto do `helloHandlerSpan`
	_, span := tracer.Start(ctx, "GetTemperatureSpan")
	defer span.End()

	// ctx := context.Background()
	// ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	config, _ := configs.LoadConfig()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s:%d/temperature/%s", config.ServiceBHost, config.ServiceBPort, cep), nil)
	if err != nil {
		return "", fmt.Errorf("error creating request for service B: %w", err)
	}

	// Withot this, the request won't carry the trace context to Service B
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error during request to service B: %w", err)
	}

	defer resp.Body.Close()

	log.Println("Response from service B:", resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body response: %w", err)
	}
	log.Println("Response body from service B:", string(body))

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("service B returned status code %d with body: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func checkCep(cep string) (bool, error) {
	if len(cep) != 8 {
		return false, fmt.Errorf("invalid zipcode")
	}
	return true, nil
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conn, err := initConn()
	if err != nil {
		log.Fatal(err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			serviceName,
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	shutdownTracerProvider, err := initTracerProvider(ctx, res, conn)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownTracerProvider(ctx); err != nil {
			log.Fatalf("failed to shutdown TracerProvider: %s", err)
		}
	}()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/temperature", handler)

	log.Println("Starting server on :8080")
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("An error occurred while starting the server: %v", err)
	}
}
