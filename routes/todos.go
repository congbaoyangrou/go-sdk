package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/leancloud/go-sdk/leancloud"
)

type Todo struct {
	Content string `json:"content`
}

var client *leancloud.Client

func init() {
	client = leancloud.NewEnvClient()
}

func GetTodos(c echo.Context) error {
	objects, err := client.Class("Todo").NewQuery().Order("createdAt").Find()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	todos := make([]Todo, len(objects))

	for i := 0; i < len(todos); i++ {
		objects[i].ToStruct(&(todos[i]))
	}

	return c.Render(http.StatusOK, "todos", struct {
		Title string
		Todos []Todo
	}{
		Title: "TODO 列表",
		Todos: todos,
	})
}

func PostTodos(c echo.Context) error {
	content := c.FormValue("content")
	todo := Todo{
		Content: content,
	}

	if _, err := client.Class("Todo").Create(todo); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusSeeOther, "/todos")
}
