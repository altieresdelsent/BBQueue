package main

import (
	"fmt"
	"github.com/altieresdelsent/BBQueue/queue"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	bbQueue       *queue.ExtraQueue
	redisInstance *queue.RedisInFlightStorageAndQueue
)

const (
	contentTypeJSON = "application/json; charset=utf-8"
	TimeoutKey      = "Timeout-Key"
	QueueKey        = "Queue-Key"
)

func init() {
	redisInstance = &queue.RedisInFlightStorageAndQueue{}
	err := redisInstance.Init()
	if err != nil {
		fmt.Println("Error starting redis")
		return
	}
	bbQueue = queue.NewExtraQueue(redisInstance, redisInstance)
	go func() {
		for {
			bbQueue.ReAddExpiredKeys()
			time.Sleep(time.Second)
		}
	}()

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
	server.GET("/queue/:key", serverGetKey)
}

func serverGet(c *gin.Context) {
	timeoutString := c.Request.Header.Get(TimeoutKey)
	timeoutSec, err := strconv.Atoi(timeoutString)
	if err != nil {
		c.Data(http.StatusBadRequest, contentTypeJSON, []byte(""))
		return
	}

	key, message, empty := bbQueue.Get(time.Duration(timeoutSec) * time.Second)
	if empty {
		c.Data(http.StatusNoContent, contentTypeJSON, []byte(""))
		return
	}

	c.Header("Queue-Key", key.String())
	c.Data(http.StatusOK, contentTypeJSON, []byte(message))
}

func serverPost(c *gin.Context) {
	message, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
	err = bbQueue.Add(string(message))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
		return
	}
}

func serverDelete(c *gin.Context) {
	key, err := uuid.Parse(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	err = bbQueue.Delete(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func serverGetKey(c *gin.Context) {
	key, err := uuid.Parse(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
	message, timeout, empty, err := bbQueue.GetInProcessing(key)
	if empty {
		c.Data(http.StatusNoContent, contentTypeJSON, []byte(""))
		return
	}
	c.Header(QueueKey, key.String())

	c.Header(TimeoutKey, timeout.String())
	c.Data(http.StatusOK, contentTypeJSON, []byte(message))
}
