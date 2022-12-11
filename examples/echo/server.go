package main

import (
	"github.com/gin-gonic/gin"
	"github.com/mono424/go-pts"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	gorilla "github.com/mono424/go-pts-gorilla-connector"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	r := gin.Default()

	r.Static("js/", "html/node_modules/go-pts-client/dist/")
	r.LoadHTMLGlob("html/*.html")

	tubeSystem := pts.New(gorilla.NewConnector(
		websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		func(err *pts.Error) {
			println(err.Description)
		},
	))

	tubeSystem.RegisterChannel("echo", pts.ChannelHandlers{
		OnSubscribe: func(s *pts.Context) {
			print("Client joined: " + s.FullPath)
		},
		OnMessage: func(s *pts.Context, message *pts.Message) {
			print("New Message on " + s.FullPath + ": " + string(message.Payload))
			s.Send(message.Payload)
		},
		OnUnsubscribe: func(s *pts.Context) {
			print("Client left: " + s.FullPath)
		},
	})

	r.Use(func(c *gin.Context) {
		c.Set("tubeSystem", tubeSystem)
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"socketUrl": "ws://localhost:" + port + "/connect",
		})
	})

	r.GET("/connect", func(c *gin.Context) {
		properties := make(map[string]interface{}, 1)
		properties["ctx"] = c

		if err := tubeSystem.HandleRequest(c.Writer, c.Request, properties); err != nil {
			println("Something went wrong while handling a Socket request")
		}
	})

	if err := r.Run(":" + port); err != nil {
		panic("Failed to start server")
	}
}
