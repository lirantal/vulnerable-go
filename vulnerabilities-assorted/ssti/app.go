package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func search(c *gin.Context) {
	query := c.Query("q")

	db, err := sql.Open("sqlite3", "posts.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(fmt.Sprintf("SELECT * FROM posts WHERE title LIKE '%%%s%%'", query))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error executing query")
		return
	}
	defer rows.Close()

	var results []struct {
		ID      int
		Title   string
		Content string
	}

	for rows.Next() {
		var id int
		var title string
		var content string
		err := rows.Scan(&id, &title, &content)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, struct {
			ID      int
			Title   string
			Content string
		}{id, title, content})
	}

	tmpl, err := template.New("search").Parse(fmt.Sprintf(`
                <h2>Results for query "%s":</h2>
                <ul>
                {{range .Results}}
                <li><a href="{{.Content}}">{{.Title}}</a></li>
                {{end}}
                </ul>
        `, query))
	if err != nil {
		c.String(http.StatusInternalServerError, "Error creating template")
		return
	}

	data := struct {
		Query   string
		Results []struct {
			ID      int
			Title   string
			Content string
		}
	}{
		Query:   query,
		Results: results,
	}

	// Set the content type to text/html and render the template
	c.Header("Content-Type", "text/html")

	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error executing template")
		return
	}
}

func main() {
	r := gin.Default()
	r.GET("/search", search)
	r.Run(":8080")
}
