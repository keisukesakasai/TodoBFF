package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func StartMainServer() {
	log.Println("info: Start Server" + "port: " + serverPort)

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
	r.LoadHTMLGlob(pathStatic + "/templates/*")
	r.Static("/static/", pathStatic)

	//--- handler 設定
	r.GET("/", top)
	r.GET("/login", getLogin)
	r.POST("/login", postLogin)
	r.GET("/signup", getSignup)
	r.POST("/signup", postSignup)

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

	r.GET("/logout", getLogout)

	r.Run(":" + serverPort)
}

func checkSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "セッションチェック開始")
		defer span.End()
		Logger(c, "セッションチェック開始", span)
		// log.Println("セッションチェック開始")

		session := sessions.Default(c)
		LoginInfo.UserID = session.Get("UserId")

		if LoginInfo.UserID == nil {
			log.Println(LoginInfo.UserID.(string) + " はログインしていません")
			c.Redirect(http.StatusMovedPermanently, "/login")
			c.Abort()
		} else {
			log.Println(LoginInfo.UserID.(string) + " をセッション ID にセットしました")
			c.Set("UserId", LoginInfo.UserID) // ユーザIDをセット
			c.Next()
		}

		_, span = tracer.Start(c.Request.Context(), "セッションチェック終了")
		defer span.End()
		Logger(c, "セッションチェック終了", span)
		// log.Println("セッションチェック終了")
	}
}
