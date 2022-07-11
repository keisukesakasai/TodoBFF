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

type ResponseGetUser struct {
	ID        int    `json:"ID"`
	UUID      string `json:"UUID"`
	Name      string `json:"Name"`
	Email     string `json:"Email"`
	PassWord  string `json:"PassWord"`
	CreatedAt string `json:"CreatedAt"`
}

type ResponseEncrypt struct {
	PassWord string `json:"PassWord"`
}

func getSignup(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ユーザ登録画面取得")
	defer span.End()

	log.Println("ユーザ登録画面取得")
	generateHTML(c, nil, "signup", "layout", "signup", "public_navbar")
}

func postSignup(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ユーザ登録")
	defer span.End()

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	// UserAPI createUser rpc 実行
	name := c.Request.PostFormValue("name")
	email := c.Request.PostFormValue("email")
	password := c.Request.PostFormValue("password")

	jsonStr := `{"Name":"` + name + `",
	"Email":"` + email + `",
	"PassWord":"` + password + `"}`

	rsp, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/createUser",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp.Body.Close()

	body, _ := ioutil.ReadAll(rsp.Body)
	fmt.Println(string(body))

	UserId := email
	log.Println("ログイン処理")
	login(c, UserId)

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()

	log.Println("TODO画面にリダイレクト")
	c.Redirect(http.StatusMovedPermanently, "/menu/todos")
}

func getLogin(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログイン画面取得")
	defer span.End()

	log.Println("ログイン画面取得")
	generateHTML(c, nil, "login", "layout", "login", "public_navbar")
}

func postLogin(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログイン")
	defer span.End()

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	// UserAPI getUserByEmail rpc 実行
	email := c.Request.PostFormValue("email")
	jsonStr := `{"Email":"` + email + `"}`

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
	log.Println(responseGetUser)

	// UserAPI encrypt rpc 実行
	password := c.Request.PostFormValue("password")
	jsonStr = `{"PassWord":"` + password + `"}`

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
		log.Println(err)
		log.Println("ユーザがいません")
		c.Redirect(http.StatusFound, "/login")
	} else if responseEncrypt.PassWord == responseGetUser.PassWord {
		UserId := c.PostForm("email")
		log.Println("ログイン処理")
		login(c, UserId)
		c.SetCookie("UserId", UserId, 60, "/", "localhost", false, true)
		c.Redirect(http.StatusMovedPermanently, "/menu/todos")
	} else {
		log.Println("PW が間違っています")
		c.Redirect(http.StatusFound, "/login")
	}
}

func getLogout(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログアウト")
	defer span.End()

	logout(c)

	_, span = tracer.Start(c.Request.Context(), "TOP画面にリダイレクト")
	defer span.End()

	log.Println("TOP画面にリダイレクト")
	c.Redirect(http.StatusMovedPermanently, "/")
}

func login(c *gin.Context, UserId string) {
	_, span := tracer.Start(c.Request.Context(), "ログイン処理...")
	defer span.End()

	session := sessions.Default(c)
	session.Set("UserId", UserId)
	session.Save()
	log.Println("ログイン")
}

func logout(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "ログアウト処理...")
	defer span.End()

	session := sessions.Default(c)
	session.Clear()
	session.Save()
	log.Println("ログアウト")
}
