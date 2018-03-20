package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

const REDIS_WEIGHTS_KEY = "DWR_WEIGHTS"

type WeightsBundle struct {
	sync.RWMutex
	uw map[string]int32 // userWeights
	dw map[string]int32 // deducedWeights
}

type Weights map[string]WeightsBundle

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("key", val)

	// we check in redis if we already have some data

}

func run_web() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/add-weights", addWeights)
	e.PUT("/update-weights/:key", updateWeights)
	e.GET("/:key", getValue)
	e.GET("/hello", hello)

	// Start server
	e.Logger.Fatal(e.Start(":30300"))
}

// Handlers
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func addWeights(c echo.Context) error {
	return c.String(http.StatusOK, "okay")
}

func updateWeights(c echo.Context) error {
	return c.String(http.StatusOK, "okay")
}

func getValue(c echo.Context) error { // for lack of a better name
	key := c.Param("key")
	return c.String(http.StatusOK, key)
}

func main() {
	run_web()
}
