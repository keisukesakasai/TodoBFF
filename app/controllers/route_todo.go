package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type User struct {
	ID        int
	UUID      string
	Name      string
	Email     string
	PassWord  string
	CreatedAt time.Time
	Todos     []Todo
}

type Todo struct {
	ID        int
	Content   string
	UserID    int
	CreatedAt time.Time
}

type Todos struct {
	Todos []Todo
}

type getTodosByUserResponse struct {
	Todos []Todo `json:"todos"`
}

type getTodoResponse struct {
	ID        int       `json:"ID"`
	Content   string    `json:"Content"`
	UserID    int       `json:"UserID"`
	CreatedAt time.Time `json:"CreatedAt"`
}

type updateTodoResponse struct {
	Content string `json:"Content"`
}

type deleteTodoResponse struct {
	ResultCode string `json:"resultCode`
}

func top(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TOP画面取得")
	defer span.End()

	log.Println("TOP画面取得")
	generateHTML(c, "hello", "top", "layout", "top", "public_navbar")
}

func getIndex(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "TODO画面取得")
	defer span.End()

	UserId, isExist := c.Get("UserId")
	if !isExist {
		log.Println("セッションが存在していません")
	}

	// UserAPI getUserByEmail rpc 実行
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

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
	log.Println(responseGetUser)

	// TodoAPI getTodosByUser rpc 実行
	user_id := strconv.Itoa(responseGetUser.ID)
	jsonStr2 := `{"user_id":"` + string(user_id) + `"}`

	log.Println("user_id が " + user_id + " の Todo を参照")

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/getTodosByUser",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	log.Println(rsp2)
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
	log.Println("TODO画面取得")

	generateHTML(c, user, "index", "layout", "private_navbar", "index")
	// generateHTML(c, user, "index", "layout", "private_navbar", "index")
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

	UserId, isExist := c.Get("UserId")
	if !isExist {
		log.Println("セッションが存在していません")
	}

	//--- UserAPI getUserByEmail rpc 実行
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

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

	//--- TodoAPI getTodosByUser rpc 実行
	log.Println("---responseGetUser---")
	log.Println(responseGetUser.ID)
	user_id := strconv.Itoa(responseGetUser.ID)
	content := c.Request.PostFormValue("content")

	jsonStr2 := `{"Content":"` + content + `",
	"User_Id":"` + user_id + `"}`

	log.Println(jsonStr2)

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/createTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	log.Println("---")
	log.Println(rsp2)
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
	log.Println("TODO保存")

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()

	log.Println("TODO画面にリダイレクト")
	c.Redirect(http.StatusFound, "/menu/todos")
}

func getTodoEdit(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO編集画面取得")
	defer span.End()

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	UserId, _ := c.Get("UserId")
	//--- UserAPI getUserByEmail rpc 実行
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

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

	//--- TodoAPI getTodo rpc 実行
	todo_id := strconv.Itoa(id)
	jsonStr2 := `{"todo_id":"` + todo_id + `"}`

	log.Println(jsonStr2)

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/getTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	log.Println("---")
	log.Println(rsp2)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	log.Println(byteArr)
	var getTodoresponse getTodoResponse
	err = json.Unmarshal(byteArr, &getTodoresponse)
	if err != nil {
		log.Println(err)
	}
	log.Println("TODO参照")

	log.Println("TODO編集画面取得")
	generateHTML(c, getTodoresponse, "todoEdit", "layout", "private_navbar", "todo_edit")
}

func postTodoUpdate(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO更新")
	defer span.End()

	err := c.Request.ParseForm()
	if err != nil {
		log.Println(err)
	}

	UserId, _ := c.Get("UserId")
	//--- UserAPI getUserByEmail rpc 実行
	email := UserId.(string)
	jsonStr1 := `{"Email":"` + email + `"}`

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

	//--- TodoAPI updateTodo rpc 実行
	content := c.Request.PostFormValue("content")
	user_id := strconv.Itoa(responseGetUser.ID)
	todo_id := strconv.Itoa(id)
	jsonStr2 := `{"Content":"` + content + `",
	"User_Id":"` + user_id + `",
	"Todo_Id":"` + todo_id + `"}`

	log.Println(jsonStr2)

	rsp2, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/updateTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr2)),
	)
	log.Println("---")
	log.Println(rsp2)
	if err != nil {
		log.Println(err)
		return
	}
	defer rsp2.Body.Close()

	byteArr, _ = ioutil.ReadAll(rsp2.Body)
	log.Println(byteArr)
	var updateTodoresponse updateTodoResponse
	err = json.Unmarshal(byteArr, &updateTodoresponse)
	if err != nil {
		log.Println(err)
	}
	log.Println(updateTodoresponse.Content + "に TODO を更新しました")

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()

	log.Println("TODO画面にリダイレクト")
	c.Redirect(http.StatusFound, "/menu/todos")
}

func getTodoDelete(c *gin.Context, id int) {
	_, span := tracer.Start(c.Request.Context(), "TODO削除")
	defer span.End()

	//--- TodoAPI deleteTodo rpc 実行
	todo_id := strconv.Itoa(id)
	jsonStr1 := `{"todo_id":"` + todo_id + `"}`

	log.Println(jsonStr1)

	rsp1, err := otelhttp.Post(
		c.Request.Context(),
		EpTodoAPI+"/deleteTodo",
		"application/json",
		bytes.NewBuffer([]byte(jsonStr1)),
	)
	log.Println("---")
	log.Println(rsp1)
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
	log.Println(deleteTodoresponse.ResultCode)
	log.Println("TODO削除")

	_, span = tracer.Start(c.Request.Context(), "TODO画面にリダイレクト")
	defer span.End()

	log.Println("TODO画面にリダイレクト")
	c.Redirect(http.StatusFound, "/menu/todos")
}
