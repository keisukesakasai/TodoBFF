package controllers

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"text/template"
	"time"
	"todobff/config"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("TodoBFF"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var tracerProvider *sdktrace.TracerProvider

	conn, err := grpc.DialContext(ctx, "otel-collector-collector.tracing.svc.cluster.local:4318", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// idg := xray.NewIDGenerator()

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		// sdktrace.WithIDGenerator(idg),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func generateHTML(c *gin.Context, data interface{}, procname string, filenames ...string) {
	_, span := tracer.Start(c.Request.Context(), "generateHTML : "+procname)
	defer span.End()

	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("app/views/templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(c.Writer, "layout", data)
}

var tracer = otel.Tracer("TodoBFF")

func StartMainServer() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Otel Collecotor への接続設定
	shutdown, err := initProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	// router 設定
	r := gin.New()

	r.Use(otelgin.Middleware("TodoBFF-server"))

	// template 設定
	r.LoadHTMLGlob(config.Config.Static + "/templates/*")
	r.Static("/static/", config.Config.Static)

	//--- handler 設定
	// r.GET("/", top)

	r.GET("/signup", getSignup)
	r.POST("/signup", postSignup)

	r.Run(":" + config.Config.Port)

}
