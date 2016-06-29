package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func main() {
	db := ConnectDB("localhost:6379")

	r := gin.Default()
	r.LoadHTMLGlob("*.html")

	r.Any("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/default")
	})
	r.GET("/:name", func(c *gin.Context) {
		uuids, err := db.List(c.Param("name"))
		if err != nil {
			panic(err)
		}

		notes := make([]*Post, len(uuids))
		for i, uuid := range uuids {
			notes[i], err = db.Get(uuid)
			if err != nil {
				panic(err)
			}
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"name":  c.Param("name"),
			"notes": notes,
		})
	})
	r.POST("/:name", func(c *gin.Context) {
		id, err := strconv.Atoi(c.PostForm("id"))
		if err != nil {
			panic(err)
		}

		post := &Post{
			Name: c.PostForm("name"),
			ID:   id,
			Text: c.PostForm("text"),
		}

		if _, err := db.Post(c.Param("name"), post); err != nil {
			panic(err)
		}
		c.Redirect(http.StatusSeeOther, c.Request.URL.String())
	})
	panic(r.Run())
}
