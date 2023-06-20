package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Website struct {
	app *gin.Engine
}

func NewWebsite(app *gin.Engine) *Website {
	return &Website{app: app}
}

func (w *Website) RegisterRoutes() {
	w.app.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/chat")
	})
	w.app.GET("/chat", w.index)
	w.app.GET("/chat/:conversation_id", w.chat)
	w.app.GET("/assets/:folder/:file", w.assets)
}

func (w *Website) chat(c *gin.Context) {
	conversationID := c.Param("conversation_id")
	if !strings.Contains(conversationID, "-") {
		c.Redirect(http.StatusMovedPermanently, "/chat")
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{"chat_id": conversationID})
}

func (w *Website) index(c *gin.Context) {
	chatID := fmt.Sprintf("%s-%s-%s-%s-%s", randomHex(4), randomHex(2), randomHex(2), randomHex(2), strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 16))
	c.HTML(http.StatusOK, "index.html", gin.H{"chat_id": chatID})
}

func (w *Website) assets(c *gin.Context) {
	folder := c.Param("folder")
	file := c.Param("file")
	c.File(fmt.Sprintf("./../client/%s/%s", folder, file))
}

func randomHex(n int) string {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func main() {
	app := gin.Default()
	website := NewWebsite(app)
	website.RegisterRoutes()
	app.Run()
}