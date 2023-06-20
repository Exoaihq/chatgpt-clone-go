package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SiteConfig struct {
	Port string `json:"port"`
}

type Config struct {
	SiteConfig SiteConfig `json:"site_config"`
}

type Website struct {
	app    *gin.Engine
	routes map[string]Route
}

type BackendAPI struct {
	app    *gin.Engine
	routes map[string]Route
}

type Route struct {
	Function gin.HandlerFunc
	Methods  []string
}

func main() {
	app := gin.Default()

	config := loadConfig("config.json")
	siteConfig := config.SiteConfig

	site := NewWebsite(app)
	for route, routeInfo := range site.routes {
		app.Handle(strings.Join(routeInfo.Methods, ","), route, routeInfo.Function)
	}

	backendAPI := NewBackendAPI(app, config)
	for route, routeInfo := range backendAPI.routes {
		app.Handle(strings.Join(routeInfo.Methods, ","), route, routeInfo.Function)
	}

	fmt.Printf("Running on port %s\n", siteConfig.Port)
	app.Run(":" + siteConfig.Port)
	fmt.Printf("Closing port %s\n", siteConfig.Port)
}

func loadConfig(filename string) Config {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func NewWebsite(app *gin.Engine) *Website {
	// Initialize your website routes here
	routes := make(map[string]Route)
	return &Website{app: app, routes: routes}
}

func NewBackendAPI(app *gin.Engine, config Config) *BackendAPI {
	// Initialize your backend API routes here
	routes := make(map[string]Route)
	return &BackendAPI{app: app, routes: routes}
}
