package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
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
	Threading  string `json:"threading"`
	Selector   string `json:"selector"`
	ResultMode string `json:"resultMode"`
	TopN       int    `json:"topN"`
	NodeId1    *int   `json:"nodeId1"`
	NodeId2    *int   `json:"nodeId2"`
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
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unknown mode: %s", req.Mode)})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	start := time.Now()
	logs := []src.LogEntry{}
	visitedCount := 0
	var searchErr error

	multi := req.Threading != "single"

	switch req.Algorithm {
	case "dfs":
		if multi {
			_, logs, visitedCount, searchErr = src.SearchDFS(root, req.Selector, req.TopN)
		} else {
			_, logs, visitedCount, searchErr = src.SearchDFSSingle(root, req.Selector, req.TopN)
		}

	case "bfs":
		if multi {
			_, logs, visitedCount, searchErr = src.BFSSearch(root, req.Selector, req.TopN)
		} else {
			_, logs, visitedCount, searchErr = src.BFSSearchSingle(root, req.Selector, req.TopN)
		}

	case "lca":
		if req.NodeId1 == nil && req.NodeId2 == nil {
			logs = []src.LogEntry{}
			visitedCount = 0
			break
		}

		if req.NodeId1 == nil || req.NodeId2 == nil {
			searchErr = fmt.Errorf("nodeId1 and nodeId2 are required for LCA search")
			break
		}

		lca, err := src.PreproccessLCABinaryLifting(root)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, logs, visitedCount, searchErr = lca.SearchLCAByID(*req.NodeId1, *req.NodeId2)

	default:
		searchErr = fmt.Errorf("unknown algorithm: %s", req.Algorithm)
	}

	if searchErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": searchErr.Error()})
		return
	}

	if logs == nil {
		logs = []src.LogEntry{}
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

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

// JSON Outputnya:
// {
//   "executionTimeMs": 2.5,
//   "nodesVisited": 15,
//   "maxDepth": 4,

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

//   "traversalLog": [
//      { "nodeId": 1, "tag": "html", "status": "visited" },
//      { "nodeId": 2, "tag": "body", "status": "visited" },
//      { "nodeId": 5, "tag": "div",  "status": "matched" } // <-- Marked as affected!
//   ]
// }

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/api/data", getResult)

	port := getEnv("PORT", "8080")
	fmt.Printf("Gin backend server is listening on http://localhost:%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start Gin server: ", err)
	}
}