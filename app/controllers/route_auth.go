package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func getSignup(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ユーザ登録画面取得")
	defer span.End()
	Logger(c, "ユーザ登録画面取得", span)
	// log.Println("ユーザ登録画面取得")

	generateHTML(c, nil, "signup", "layout", "signup", "public_navbar")
}

func postSignup(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ユーザ登録")
	defer span.End()
	Logger(c, "ユーザ登録", span)
	// log.Println("ユーザ登録")

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	//--- UserAPI createUser への Post
	name := c.Request.PostFormValue("name")
	email := c.Request.PostFormValue("email")
	password := c.Request.PostFormValue("password")

	jsonStr := `{"Name":"` + name + `",
	"Email":"` + email + `",
	"PassWord":"` + password + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /createUser にポスト")
	defer span.End()
	Logger(c, "UserAPI /createUser にポスト", span)
	// log.Println("UserAPI /createUser にポスト")

	rsp, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/createUser",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp.Body.Close()
	body, _ := ioutil.ReadAll(rsp.Body)
	log.Println(string(body))

	UserId := email
	login(c, UserId)

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()
	Logger(c, "TODO画面にリダイレクト", span)
	// log.Println("TODO画面にリダイレクト")

	c.Redirect(http.StatusMovedPermanently, "/menu/todos")
}

func getLogin(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログイン画面取得")
	defer span.End()
	Logger(c, "ログイン画面取得", span)
	// log.Println("ログイン画面取得")

	generateHTML(c, nil, "login", "layout", "login", "public_navbar")
}

func postLogin(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログイン")
	defer span.End()
	Logger(c, "ログイン", span)
	// log.Println("ログイン")

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	//--- UserAPI getUserByEmail への Post
	email := c.Request.PostFormValue("email")
	jsonStr := `{"Email":"` + email + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /getUserByEmail にポスト")
	defer span.End()
	Logger(c, "UserAPI /getUserByEmail にポスト", span)
	// log.Println("UserAPI /getUserByEmail にポスト")

	rsp, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/getUserByEmail",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp.Body)
	var responseGetUser ResponseGetUser
	err = json.Unmarshal(byteArr, &responseGetUser)
	if err != nil {
		log.Println(err)
	}

	//--- UserAPI encrypt への Post
	password := c.Request.PostFormValue("password")
	jsonStr = `{"PassWord":"` + password + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /encrypt にポスト")
	defer span.End()
	Logger(c, "UserAPI /encrypt にポスト", span)
	// log.Println("UserAPI /encrypt にポスト")

	rsp, err = otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/encrypt",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp.Body)
	var responseEncrypt ResponseEncrypt
	err = json.Unmarshal(byteArr, &responseEncrypt)
	if err != nil {
		log.Println(err)
	}

	if responseGetUser.ID == 0 {
		log.Println("ユーザがいません")

		_, span = tracer.Start(c.Request.Context(), "ログイン画面にリダイレクト")
		defer span.End()
		Logger(c, "ログイン画面にリダイレクト", span)
		// log.Println("ログイン画面にリダイレクト")

		c.Redirect(http.StatusFound, "/login")
	} else if responseEncrypt.PassWord == responseGetUser.PassWord {
		UserId := c.PostForm("email")
		login(c, UserId)

		_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
		defer span.End()
		Logger(c, "TODO画面にリダイレクト", span)
		// log.Println("TODO画面にリダイレクト")

		c.Redirect(http.StatusMovedPermanently, "/menu/todos")
	} else {
		log.Println("PW が間違っています")

		_, span = tracer.Start(c.Request.Context(), "ログイン画面にリダイレクト")
		defer span.End()
		Logger(c, "ログイン画面にリダイレクト", span)
		// log.Println("ログイン画面にリダイレクト")

		c.Redirect(http.StatusFound, "/login")
	}
}

func getLogout(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログアウト")
	defer span.End()

	logout(c)

	_, span = tracer.Start(c.Request.Context(), "TOP画面にリダイレクト")
	defer span.End()
	Logger(c, "TOP画面にリダイレクト", span)
	// log.Println("TOP画面にリダイレクト")

	c.Redirect(http.StatusMovedPermanently, "/")
}

func login(c *gin.Context, UserId string) {
	_, span := tracer.Start(c.Request.Context(), "ログイン処理...")
	defer span.End()
	Logger(c, "ログイン処理...", span)
	// log.Println("ログイン処理...")

	session := sessions.Default(c)

	_, span = tracer.Start(c.Request.Context(), "セッション設定")
	defer span.End()
	Logger(c, "セッション設定", span)
	// log.Println("セッション設定")

	session.Set("UserId", UserId)

	_, span = tracer.Start(c.Request.Context(), "セッション保存")
	defer span.End()
	Logger(c, "セッション保存", span)
	// log.Println("セッション保存")

	session.Save()

	log.Println("ログイン完了")
}

func logout(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログアウト処理...")
	defer span.End()
	Logger(c, "ログアウト処理...", span)
	// log.Println("ログアウト処理...")

	session := sessions.Default(c)

	_, span = tracer.Start(c.Request.Context(), "セッションクリア")
	defer span.End()
	Logger(c, "セッションクリア", span)
	// log.Println("セッションクリア")

	session.Clear()

	_, span = tracer.Start(c.Request.Context(), "セッション保存")
	defer span.End()
	Logger(c, "セッション保存", span)
	// log.Println("セッション保存")

	session.Save()

	log.Println("ログアウト完了")
}
