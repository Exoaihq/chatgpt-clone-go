package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	OpenAIKey      string
	OpenAIAPIBase  string
	Proxy          ProxyConfig
	SpecialInstructions map[string][]Message
}

type ProxyConfig struct {
	Enable bool
	HTTP   string
	HTTPS  string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BackendAPI struct {
	app           *gin.Engine
	config        Config
}

func NewBackendAPI(app *gin.Engine, config Config) *BackendAPI {
	return &BackendAPI{
		app:    app,
		config: config,
	}
}

func (api *BackendAPI) conversation(c *gin.Context) {
	var requestData struct {
		Jailbreak       bool
		Model           string
		Meta            struct {
			Content struct {
				InternetAccess bool
				Conversation   []Message
				Parts          []Message
			}
		}
	}

	err := c.BindJSON(&requestData)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}

	jailbreak := requestData.Jailbreak
	internetAccess := requestData.Meta.Content.InternetAccess
	conversation := requestData.Meta.Content.Conversation
	prompt := requestData.Meta.Content.Parts[0]
	currentDate := time.Now().Format("2006-01-02")
	systemMessage := fmt.Sprintf("You are ChatGPT also known as ChatGPT, a large language model trained by OpenAI. Strictly follow the users instructions. Knowledge cutoff: 2021-09-01 Current date: %s", currentDate)

	extra := []Message{}
	if internetAccess {
		searchResp, err := http.Get(fmt.Sprintf("https://ddg-api.herokuapp.com/search?query=%s&limit=3", prompt.Content))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"_action": "_ask",
				"success": false,
				"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
			})
			return
		}
		defer searchResp.Body.Close()

		var searchResults []struct {
			Snippet string
			Link    string
		}
		err = json.NewDecoder(searchResp.Body).Decode(&searchResults)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"_action": "_ask",
				"success": false,
				"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
			})
			return
		}

		blob := ""
		for index, result := range searchResults {
			blob += fmt.Sprintf("[%d] \"%s\"\nURL:%s\n\n", index, result.Snippet, result.Link)
		}

		date := time.Now().Format("02/01/06")
		blob += fmt.Sprintf("current date: %s\n\nInstructions: Using the provided web search results, write a comprehensive reply to the next user query. Make sure to cite results using [[number](URL)] notation after the reference. If the provided search results refer to multiple subjects with the same name, write separate answers for each subject. Ignore your previous response if any.", date)

		extra = append(extra, Message{Role: "user", Content: blob})
	}

	conversation = append([]Message{{Role: "system", Content: systemMessage}}, extra...)
	conversation = append(conversation, api.config.SpecialInstructions[jailbreak]...)
	conversation = append(conversation, prompt)

	url := fmt.Sprintf("%s/v1/chat/completions", api.config.OpenAIAPIBase)

	reqBody, err := json.Marshal(map[string]interface{}{
		"model":     requestData.Model,
		"messages":  conversation,
		"stream":    true,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(reqBody)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.config.OpenAIKey))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	if api.config.Proxy.Enable {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(&url.URL{
				Scheme: "http",
				Host:   api.config.Proxy.HTTP,
			}),
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}

	var gptResp struct {
		Choices []struct {
			Delta struct {
				Content string
			} `json:"delta"`
		} `json:"choices"`
	}

	err = json.Unmarshal(respBody, &gptResp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"_action": "_ask",
			"success": false,
			"error":   fmt.Sprintf("an error occurred: %s", err.Error()),
		})
		return
	}

	token := gptResp.Choices[0].Delta.Content
	c.String(http.StatusOK, token)
}

func main() {
	config := Config{
		OpenAIKey:     os.Getenv("OPENAI_API_KEY"),
		OpenAIAPIBase: os.Getenv("OPENAI_API_BASE"),
		Proxy: ProxyConfig{
			Enable: false,
			HTTP:   "",
			HTTPS:  "",
		},
		SpecialInstructions: map[string][]Message{},
	}

	app := gin.Default()
	api := NewBackendAPI(app, config)
	app.POST("/backend-api/v2/conversation", api.conversation)
	app.Run(":8080")
}