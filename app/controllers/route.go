package controllers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

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

	// URLを生成
	u := &url.URL{}
	u.Scheme = "http"
	u.Host = "localhost:8090"
	u.Path = "/signup"
	// url文字列
	uStr := u.String()

	// ポストデータ
	name := c.Request.PostFormValue("name")
	email := c.Request.PostFormValue("email")
	password := c.Request.PostFormValue("password")
	jsonStr := `{"Name":"` + name + `",
	"Email":"` + email + `",
	"PassWord":"` + password + `"}`

	rsp, err := http.Post(uStr,
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
}
