package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func top(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TOP画面取得")
	defer span.End()
	log.Println("TOP画面取得")

	generateHTML(c, "hello", "top", "layout", "top", "public_navbar")
}

func getIndex(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TODO画面取得")
	defer span.End()
	log.Println("TODO画面取得")

	UserId, isExist := c.Get("UserId")
	if !isExist {
		log.Println("セッションが存在していません")
	}

	//--- UserAPI getUserByEmail への Post
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /getUserByEmail にポスト")
	defer span.End()
	log.Println("UserAPI /getUserByEmail にポスト")

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/getUserByEmail",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp1.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp1.Body)
	var responseGetUser ResponseGetUser
	err = json.Unmarshal(byteArr, &responseGetUser)
	if err != nil {
		log.Println(err)
	}

	//--- TodoAPI getTodosByUser への Post
	user_id := strconv.Itoa(responseGetUser.ID)
	jsonStr2 := `{"user_id":"` + string(user_id) + `"}`

	_, span = tracer.Start(c.Request.Context(), "TodoAPI /getTodosByEmail にポスト")
	defer span.End()
	log.Println("TodoAPI /getTodosByEmail にポスト")

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/getTodosByUser",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	var getTodosByUserresponse getTodosByUserResponse
	err = json.Unmarshal(byteArr, &getTodosByUserresponse)
	if err != nil {
		log.Println(err)
	}

	var user User
	user.Name = responseGetUser.Name
	user.Todos = getTodosByUserresponse.Todos

	_, span = tracer.Start(c.Request.Context(), "TODO画面取得")
	defer span.End()
	log.Println("TODO画面取得")

	generateHTML(c, user, "index", "layout", "private_navbar", "index")
}

func getTodoNew(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TODO作成画面取得")
	defer span.End()
	log.Println("TODO作成画面取得")

	generateHTML(c, nil, "todoNew", "layout", "private_navbar", "todo_new")
}

func postTodoSave(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TODO保存")
	defer span.End()
	log.Println("TODO保存")

	UserId, isExist := c.Get("UserId")
	if !isExist {
		log.Println("セッションが存在していません")
	}

	//--- UserAPI getUserByEmail への Post
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /getUserByEmail にポスト")
	defer span.End()
	log.Println("UserAPI /getUserByEmail にポスト")

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/getUserByEmail",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp1.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp1.Body)
	var responseGetUser ResponseGetUser
	err = json.Unmarshal(byteArr, &responseGetUser)
	if err != nil {
		log.Println(err)
	}

	//--- TodoAPI createTodo への Post
	user_id := strconv.Itoa(responseGetUser.ID)
	content := c.Request.PostFormValue("content")

	_, span = tracer.Start(c.Request.Context(), "TodoAPI /createTodo にポスト")
	defer span.End()
	log.Println("TodoAPI /createTodo にポスト")

	jsonStr2 := `{"Content":"` + content + `",
	"User_Id":"` + user_id + `"}`

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/createTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	var getTodosByUserresponse getTodosByUserResponse
	err = json.Unmarshal(byteArr, &getTodosByUserresponse)
	if err != nil {
		log.Println(err)
	}

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()
	log.Println("TODO画面にリダイレクト")

	c.Redirect(http.StatusFound, "/menu/todos")
}

func getTodoEdit(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO編集画面取得")
	defer span.End()
	log.Println("TODO編集画面取得")

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	UserId, _ := c.Get("UserId")
	//--- UserAPI getUserByEmail への Post
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /getUserByEmail にポスト")
	defer span.End()
	log.Println("UserAPI /getUserByEmail にポスト")

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/getUserByEmail",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp1.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp1.Body)
	var responseGetUser ResponseGetUser
	err = json.Unmarshal(byteArr, &responseGetUser)
	if err != nil {
		log.Println(err)
	}

	//--- TodoAPI getTodo への Post
	todo_id := strconv.Itoa(id)
	jsonStr2 := `{"todo_id":"` + todo_id + `"}`

	_, span = tracer.Start(c.Request.Context(), "TodoAPI /getTodo にポスト")
	defer span.End()
	log.Println("TodoAPI /getTodo にポスト")

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/getTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	var getTodoresponse getTodoResponse
	err = json.Unmarshal(byteArr, &getTodoresponse)
	if err != nil {
		log.Println(err)
	}

	_, span = tracer.Start(c.Request.Context(), "TODO編集画面取得")
	defer span.End()
	log.Println("TODO編集取得")

	generateHTML(c, getTodoresponse, "todoEdit", "layout", "private_navbar", "todo_edit")
}

func postTodoUpdate(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO更新")
	defer span.End()
	log.Println("TODO更新")

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	UserId, _ := c.Get("UserId")
	//--- UserAPI getUserByEmail への Post
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

	_, span = tracer.Start(c.Request.Context(), "UserAPI /getUserByEmail にポスト")
	defer span.End()
	log.Println("UserAPI /getUserByEmail にポスト")

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpUserApi+"/getUserByEmail",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rsp1.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp1.Body)
	var responseGetUser ResponseGetUser
	err = json.Unmarshal(byteArr, &responseGetUser)
	if err != nil {
		log.Println(err)
	}

	//--- TodoAPI updateTodo への Post
	content := c.Request.PostFormValue("content")
	user_id := strconv.Itoa(responseGetUser.ID)
	todo_id := strconv.Itoa(id)
	jsonStr2 := `{"Content":"` + content + `",
	"User_Id":"` + user_id + `",
	"Todo_Id":"` + todo_id + `"}`

	_, span = tracer.Start(c.Request.Context(), "TodoAPI /updateTodo にポスト")
	defer span.End()
	log.Println("TodoAPI /updateTodo にポスト")

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/updateTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	var updateTodoresponse updateTodoResponse
	err = json.Unmarshal(byteArr, &updateTodoresponse)
	if err != nil {
		log.Println(err)
	}

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()
	log.Println("TODO画面にリダイレクト")

	c.Redirect(http.StatusFound, "/menu/todos")
}

func getTodoDelete(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO削除")
	defer span.End()
	log.Println("TODO削除")

	//--- TodoAPI deleteTodo への Post
	todo_id := strconv.Itoa(id)
	jsonStr1 := `{"todo_id":"` + todo_id + `"}`

	_, span = tracer.Start(c.Request.Context(), "TodoAPI /deleteTodo にポスト")
	defer span.End()
	log.Println("TodoAPI /deleteTodo にポスト")

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/deleteTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp1.Body.Close()

	byteArr, _ := ioutil.ReadAll(rsp1.Body)
	var deleteTodoresponse deleteTodoResponse
	err = json.Unmarshal(byteArr, &deleteTodoresponse)
	if err != nil {
		log.Println(err)
	}

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()
	log.Println("TODO画面にリダイレクト")

	c.Redirect(http.StatusFound, "/menu/todos")
}
