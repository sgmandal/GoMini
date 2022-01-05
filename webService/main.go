package main

import (
	"chat2/webService/hubconn"

	"github.com/gin-gonic/gin"
)

func main() {

	hello := hubconn.NewHub()
	go hello.Run()

	router := gin.New()
	router.LoadHTMLFiles("index.html")
	router.GET("/room/:roomId", RenderChat)
	router.GET("/ws/:roomId", RenderWebSocket)
	router.Run("localhost:8080")
}

func RenderChat(c *gin.Context) {
	c.HTML(200, "index.html", nil)
}

func RenderWebSocket(c *gin.Context) {
	roomId := c.Param("roomId")
	hubconn.ServeWs(c.Writer, c.Request, roomId)
}
