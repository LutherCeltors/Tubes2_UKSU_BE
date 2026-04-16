package main

import (
	"fmt"
	"log"
	"net/http"
	"tubes2_cauksu_be/src"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type InputData struct {
	Mode       string `json:"mode"`
	Url        string `json:"url"`
	Html       string `json:"html"`
	Algorithm  string `json:"algorithm"`
	Selector   string `json:"selector"`
	ResultMode string `json:"resultMode"`
	TopN       int    `json:"topN"`
}

func getResult(c *gin.Context) {
	var req InputData
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var root *src.Node
	var err error
	switch req.Mode {
	case "url":
		root, err = src.ParseURLToDOMTree(req.Url)
	case "html":
		root, err = src.ParseToDOMTreeManual(req.Html)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// c.IndentedJSON(response)
}
func main() {
	r := gin.Default()

	// Configure CORS Middleware
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/api/data", getResult)

	port := "8080"
	fmt.Printf("Gin backend server is listening on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start Gin server: ", err)
	}
}
