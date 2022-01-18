package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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

func (op *CRUD) query(c *gin.Context) {

	// connect to DB
	conn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	db, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("Open DB failed. Error msg:", err)
		return
	}
	if err := db.Ping(); err != nil {
		fmt.Println("Connect to DB failed. Error msg:", err.Error())
		return
	}
	defer db.Close()

}

func (op *CRUD) CreateTable(db *sql.DB) error {
	/// create and initial fType table ///
	op.sql = `
		CREATE TABLE IF NOT EXISTS fType(
		ftId INT PRIMARY KEY AUTO_INCREMENT NOT NULL,
		ftName VARCHAR(32)
		);`

	if _, err := db.Exec(op.sql); err != nil {
		fmt.Println("create table fail:", err)
		return err
	}
	fmt.Println("create table succeed!")

	op.sql = `
	INSERT IGNORE INTO fType (ftId,ftName)
	VALUES(1 , 'type1'),
	(2 , 'type2'),
	(3 , 'type3');
	`
	if _, err := db.Exec(op.sql); err != nil {
		fmt.Println("insert table fail:", err)
		return err
	}
	fmt.Println("inset fType table succeed!")

	/// create and initial main fruits and reference to fType ///
	op.sql = `
		CREATE TABLE IF NOT EXISTS fruits(
		fId INT(4) PRIMARY KEY AUTO_INCREMENT NOT NULL,
		fName VARCHAR(32),
        fPrice INT,
		fNum INT,
		fType INT,
		FOREIGN KEY(fType) REFERENCES fType(ftId)
		);`

	if _, err := db.Exec(op.sql); err != nil {
		fmt.Println("create table fail:", err)
		return err
	}
	fmt.Println("create table succeed!")

	op.sql = `INSERT IGNORE INTO fruits (fid,fName,fPrice,fNum,fType)
	VALUES(1,'apple',55,3,1),
	(2,'orange',23,6,2),
	(3,'banana',33,5,3),
	(4,'lemon',123,1,1),
	(5,'mango',41,3,2),
	(6,'grape',55,3,2),
	(7,'blackberry',56,2,3),
	(8,'berry',11,9,1),
	(9,'coconut',32,4,3);`

	if _, err := db.Exec(op.sql); err != nil {
		fmt.Println("insert table fail:", err)
		return err
	}
	fmt.Println("inset fruits table succeed!")

	return nil
}

func main() {

	dbOp := new(CRUD)
	dbOp.conn = fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	db, err := sql.Open("mysql", dbOp.conn)
	if err != nil {
		fmt.Println("Open DB failed. Error msg:", err)
		return
	}
	if err := db.Ping(); err != nil {
		fmt.Println("Connect to DB failed. Error msg:", err.Error())
		return
	}
	defer db.Close()

	dbOp.CreateTable(db)

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/longJob", lagEcho) //websocket
	r.POST("/shortJob", echo)  //http
	r.POST("/readDB", (dbOp).query)

	r.Run(":7777")
}
