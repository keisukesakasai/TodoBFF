package controllers

import (
	"log"

	"github.com/gin-gonic/gin"
)

func top(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TOP画面取得")
	defer span.End()

	log.Println("TOP画面取得")
	generateHTML(c, "hello", "top", "layout", "top", "public_navbar")
}

/*
func index(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TODO画面取得")
	defer span.End()

	UserId, isExist := c.Get("UserId")
	if !isExist {
		log.Println("セッションが存在していません")
	}
	user, err := models.GetUserByEmail(c, UserId.(string))
	if err != nil {
		log.Println(err)
	}

	// ユーザの Todo を取得
	todos, _ := user.GetTodosByUser(c)
	user.Todos = todos

	log.Println("TODO画面取得")
	generateHTML(c, user, "index", "layout", "private_navbar", "index")
}
*/
