package main

import (
	"github.com/altieresdelsent/BBQueue/queue"
	"github.com/google/uuid"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

var bbqueue queue.ExtraQueue

const contentTypeJSON = "application/json; charset=utf-8"

func init() {
	bbqueue = queue.ExtraQueue{}
}
func main() {
	router := gin.Default()

	setupEndpoints(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}

func setupEndpoints(server *gin.Engine) {
	server.GET("/queue", serverGet)
	server.POST("/queue", serverPost)
	server.DELETE("/queue/:key", serverDelete)
}

func serverGet(c *gin.Context) {
	key, message := bbqueue.Get()
	c.Header("Queue-Key", key.String())
	c.Data(http.StatusOK, contentTypeJSON, []byte(message))
}

func serverPost(c *gin.Context) {
	message, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	bbqueue.Add(string(message))
}

func serverDelete(c *gin.Context) {
	key, err := uuid.Parse(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	bbqueue.Delete(key)
}
