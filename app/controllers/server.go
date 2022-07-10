package controllers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"todobff/app/SessionInfo"
	"todobff/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var LoginInfo SessionInfo.Session

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
	// ctx, cancel := context.WithTimeout(ctx, 100*time.Second)
	// defer cancel()

	var tracerProvider *sdktrace.TracerProvider

	traceExporter, _ := stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		// stdouttrace.WithWriter(os.Stderr),
		stdouttrace.WithWriter(io.Discard),
	)

	/*
		conn, err := grpc.DialContext(ctx, "otel-collector-collector.tracing.svc.cluster.local:4318", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		// Set up a trace exporter
		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}
	*/

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
var EpUserApi = config.Config.EpUserApi
var EpTodoAPI = config.Config.EpTodoApi

func StartMainServer() {
	fmt.Println("info: Start Server" + "port: " + config.Config.Port)

	// コンテキスト生成
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

	// Custom Middleware 設定
	r.Use(otelgin.Middleware("TodoBFF-server"))

	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	// template 設定
	r.LoadHTMLGlob(config.Config.Static + "/templates/*")
	r.Static("/static/", config.Config.Static)

	//--- handler 設定
	r.GET("/", top)
	r.GET("/login", getLogin)
	r.POST("/login", postLogin)

	r.GET("/signup", getSignup)
	r.POST("/signup", postSignup)

	/*
		rTodos := r.Group("/menu")
		rTodos.Use(checkSession())
		{
			rTodos.GET("/todos", getIndex)
			rTodos.GET("/todos/new", getTodoNew)
			rTodos.POST("/todos/save", postTodoSave)
			rTodos.GET("/todos/edit/:id", parseURL(getTodoEdit))
			rTodos.POST("/todos/update/:id", parseURL(postTodoUpdate))
			rTodos.GET("/todos/delete/:id", parseURL(getTodoDelete))
		}
	*/
	r.GET("/logout", getLogout)

	r.Run(":" + config.Config.Port)
}

func checkSession() gin.HandlerFunc {
	return func(c *gin.Context) {

		_, span := tracer.Start(c.Request.Context(), "セッションチェック開始")
		defer span.End()

		log.Println("セッションチェック開始")

		session := sessions.Default(c)
		LoginInfo.UserID = session.Get("UserId")

		if LoginInfo.UserID == nil {
			log.Println("ログインしていません")

			c.Redirect(http.StatusMovedPermanently, "/login")
			c.Abort()
		} else {
			c.Set("UserId", LoginInfo.UserID) // ユーザIDをセット
			c.Next()
		}

		_, span = tracer.Start(c.Request.Context(), "セッションチェック終了")
		defer span.End()

		log.Println("セッションチェック終了")
	}
}
