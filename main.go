package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
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

func CalculateMaxDepth(root *src.Node) int {
	if root == nil {
		return 0
	}
	maxChildDepth := 0
	for _, child := range root.Children {
		depth := CalculateMaxDepth(child)
		if depth > maxChildDepth {
			maxChildDepth = depth
		}
	}
	return 1 + maxChildDepth
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

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()
	var logs []src.LogEntry
	var visitedCount int
	var searchErr error

	if req.Algorithm == "dfs" {
		_, logs, visitedCount, searchErr = src.SearchDFS(root, req.Selector, req.TopN)
	} else if req.Algorithm == "bfs" {
		// BFS endpoint placeholder
		c.JSON(http.StatusNotImplemented, gin.H{"error": "BFS not implemented yet"})
		return
	} else {
		searchErr = fmt.Errorf("Unknown algorithm: %s", req.Algorithm)
	}

	if searchErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": searchErr.Error()})
		return
	}

	executionTimeMs := float64(time.Since(start).Microseconds()) / 1000.0
	maxDepth := CalculateMaxDepth(root)
	jsonTree := src.ConvertToJSONNode(root)

	c.JSON(http.StatusOK, gin.H{
		"executionTimeMs": executionTimeMs,
		"nodesVisited":    visitedCount,
		"maxDepth":        maxDepth,
		"tree":            jsonTree,
		"traversalLog":    logs,
	})
}

// JSON Outputnya:
// {
//   "executionTimeMs": 2.5,
//   "nodesVisited": 15,
//   "maxDepth": 4,

//   // 1. The Tree (Simplified safely for JSON without circular parent pointers)
//   "tree": {
//      "id": 1,
//      "tag": "html",
//      "attributes": {},
//      "children": [
//         {
//            "id": 2,
//            "tag": "body",
//            "attributes": {"class": "container"},
//            "children": [ ... ]
//         }
//      ]
//   },

//   // 2. The Unified Log (Matches marked directly in the log sequence)
//   "traversalLog": [
//      { "nodeId": 1, "tag": "html", "status": "visited" },
//      { "nodeId": 2, "tag": "body", "status": "visited" },
//      { "nodeId": 5, "tag": "div",  "status": "matched" } // <-- Marked as affected!
//   ]
// }

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

	r.POST("/api/data", getResult)

	port := "8080"
	fmt.Printf("Gin backend server is listening on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start Gin server: ", err)
	}
}
