package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	USERNAME = "nfvd"
	PASSWORD = "qct1234"
	NETWORK  = "tcp"
	SERVER   = "127.0.0.1"
	PORT     = 3306
	DATABASE = "nfvd"
)

type CRUD struct {
	conn string
	sql  string
}

// respond request body after 1 to 5 seconds
func lagEcho(c *gin.Context) {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Ignore cors domain chceck
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer func() {
		log.Println("disconnect")
		ws.Close()
	}()

	for {
		mType, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("receive: %s\n", msg)
		waitTime := rand.Int()%5 + 1
		time.Sleep(time.Duration(waitTime) * time.Second)
		sMsg := string(msg)
		sMsg += "\nsleep " + strconv.Itoa(waitTime) + " seconds"
		err = ws.WriteMessage(mType, []byte(sMsg))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func echo(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("read body fail")
		c.String(400, "read body fail\n")
	} else if len(body) == 0 {
		log.Println("empty body")
		c.String(400, "empty body\n")
	}
	c.String(200, string(body)+"\n")
}

func main() {
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/longJob", lagEcho) //websocket
	r.POST("/shortJob", echo)  //http

	r.Run(":7777")
}
