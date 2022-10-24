package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"todobff/config"

	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var tracer = otel.Tracer("TodoBFF")

func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()

	var tracerProvider *sdktrace.TracerProvider

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("TodoBFF"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	if deployEnv == "local" {
		log.Println("Deploy Mode: " + "local")
		traceExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			// stdouttrace.WithWriter(os.Stderr),
			stdouttrace.WithWriter(io.Discard),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tracerProvider := sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
			sdktrace.WithSpanProcessor(bsp),
		)
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.TraceContext{})
	}

	if deployEnv == "prod" {
		log.Println("Deploy Mode: " + "Prod")
		conn, err := grpc.DialContext(ctx, "opentelemetry-collector-collector.opentelemetry.svc.cluster.local:4318", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			fmt.Println("failed to create gRPC connection to collector: %w", err)
		}

		// Set up a trace exporter
		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		if config.Config.TraceBackend == "jaeger" {
			log.Println("TraceBackend Mode: " + "Jaeger")
			bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
				sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(bsp),
			)
		}

		if config.Config.TraceBackend == "xray" {
			log.Println("TraceBackend Mode: " + "X-Ray")
			idg := xray.NewIDGenerator()

			bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
				// sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(bsp),
				sdktrace.WithIDGenerator(idg),
				sdktrace.WithResource(newResource()),
			)
		}

		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.TraceContext{})
	}

	return tracerProvider.Shutdown, nil
}

func newResource() *resource.Resource {
	var LogGroupNames [1]string
	LogGroupNames[0] = "/aws/eks/fluentbit-cloudwatch/logs"
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.AWSLogGroupNamesKey.StringSlice(LogGroupNames[:]),
	)
}
