package controllers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"text/template"
	"todobff/app/SessionInfo"
	"todobff/config"

	"github.com/gin-gonic/gin"
)

var deployEnv = config.Config.Deploy
var serverPort = config.Config.Port
var pathStatic = config.Config.Static
var EpUserApi = config.Config.EpUserApi
var EpTodoAPI = config.Config.EpTodoApi

var LoginInfo SessionInfo.Session

func parseURL(fn func(*gin.Context, int)) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracer.Start(c.Request.Context(), "parseURL")
		defer span.End()

		fmt.Println(c.Request.URL.Path)
		q := validPath.FindStringSubmatch(c.Request.URL.Path)
		if q == nil {
			http.NotFound(c.Writer, c.Request)
			return
		}

		id, _ := strconv.Atoi(q[2])
		fmt.Println(id)
		fn(c, id)
	}
}

var validPath = regexp.MustCompile("^/menu/todos/(edit|save|update|delete)/([0-9]+)$")

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
