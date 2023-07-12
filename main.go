package main

import (
	_ "github.com/altieresdelsent/BBQueue/docs"
	"github.com/altieresdelsent/BBQueue/queue"
	"github.com/altieresdelsent/BBQueue/stressTest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100
	redisInstance = &queue.RedisInFlightStorageAndQueue{}
	err := redisInstance.Init()
	if err != nil {
		panic("Error starting redis")
		return
	}
	bbQueue = queue.NewExtraQueue(redisInstance, redisInstance)
	go func() {
		for {
			bbQueue.ReAddExpiredKeys()
			//TODO: create variable and move it to a config file
			time.Sleep(time.Second)
		}
	}()

}

// @title BBQueue API
// @version 1.0
// @description A simple queue service
// @host 127.0.0.1:8080
// @BasePath /
func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	setupEndpoints(router)
	//stressTestAllOperations()
	err := router.Run(":8080")
	if err != nil {
		return
	}
}

func stressTestAllOperations() {
	go func() {
		time.Sleep(time.Second * 3)
		stressTest.StressTestPost()
	}()

	go func() {
		time.Sleep(time.Second * 2)
		stressTest.StressTestGetAndDelete()
	}()
}

func setupEndpoints(RestApi *gin.Engine) {
	RestApi.GET("/queue", serverGet)
	RestApi.POST("/queue", serverPost)
	RestApi.DELETE("/processing/:key", serverDelete)
	RestApi.GET("/processing/:key", serverGetKey)
	RestApi.GET("/docs/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

// @summary Get the first message in the queue
// @description Returns the first message in the front of the queue, including the timeout the message can stay in processing in seconds.
// @tags Queue
// @accept json
// @produce json
// @param Timeout-Key header int true "Timeout in seconds"
// @success 200 {string} string "Ok"
// @failure 400 {string} string "Bad Request"
// @failure 204 {string} string "Queue is empty"
// @router /queue [get]
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

// @summary Add a message to the queue
// @description Adds the message contained on the body of the request to the end of the queue.
// @tags Queue
// @accept json
// @produce json
// @param message body string true "Message to add to the queue"
// @success 200 {string} string "Message added successfully"
// @failure 500 {string} string "Failed to read request body"
// @router /queue [post]
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

// @summary Delete a specific message
// @description Deletes a message with the same key as the one passed on the requested URL
// @tags Queue
// @accept json
// @produce json
// @param key path string true "Key of the message to delete"
// @success 200 {string} string "Message deleted successfully"
// @failure 204 {string} string "Key not found"
// @failure 400 {string} string "Key is not in UUID format"
// @failure 500 {string} string "Error deleting message"
// @router /processing/{key} [delete]
func serverDelete(c *gin.Context) {
	key, err := uuid.Parse(c.Param("key"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	empty, err := bbQueue.Delete(key)
	if empty {
		c.Data(http.StatusNoContent, contentTypeJSON, []byte(""))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// @summary Get a specific message
// @description Gets the message from the in-processing with the same key as the one passed on the requested URL
// @tags Queue
// @accept json
// @produce json
// @param key path string true "Key of the message to return"
// @success 200 {string} string "Ok"
// @failure 404 {string} string "Key not found"
// @failure 204 {string} string "Queue is empty"
// @router /processing/{key} [get]
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
