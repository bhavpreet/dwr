package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"

	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	config "github.com/micro/go-config"
	"github.com/micro/go-config/source/file"
)

var (
	conf    dwrConfig
	rClient *redis.Client
)

func _init() {
	// Load json config file
	config.Load(file.NewSource(
		file.WithPath("config.json"),
	))

	config.Scan(&conf)

	rClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	val, err := rClient.Get("key").Result()
	if err != nil {
		panic(err)
	}

	fmt.Println("key", val)
}

func runWeb() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/register/:key", addKeyWeights)
	e.GET("/:key", getValue)
	e.DELETE("/:key", deleteKey)

	// Start server
	e.Logger.Fatal(e.Start(":30300"))
}

// Handlers
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

type weightConstruct struct {
	Key     string         `json:"key"`
	Weights map[string]int `json:"weights"`
}

func addKeyWeights(c echo.Context) error {
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(http.StatusBadRequest, nil)
	}

	wm := make(map[string]int)
	if err := c.Bind(wm); err != nil {
		return c.JSON(http.StatusNotAcceptable, nil)
	}

	w := make(weights)
	w[key] = new(weightsBundle)
	w[key].UW = wm

	// Before we computeWeights; lets retrieve existing entry for the same
	// key if it exists and try to compare. If same, we can ignore
	if val, err := rClient.Get(
		conf.Redis.Prefix + key).Result(); err != redis.Nil {
		rW := weights{}
		if err := json.Unmarshal([]byte(val), &rW); err != nil {
			if reflect.DeepEqual(w[key].UW, rW[key].UW) {
				return c.JSON(http.StatusOK, "Same Same, Nothing Updated")
			}
		}
	}

	// Compute weights
	w[key].DW = make(kwA, 0) // Create new DW for population
	w[key].ComputeWeights()

	// Write to store
	if jSr, err := json.Marshal(w); err == nil {
		if err := rClient.Set(key, jSr, 0).Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else {
		return c.JSON(http.StatusInternalServerError, nil)
	}

	return c.JSON(http.StatusOK, "Updated")
}

func getValue(c echo.Context) error { // for lack of a better name
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(http.StatusBadRequest, nil)
	}

	var sum int
	var retKey string

	// retrieve the weights from store
	if val, err := rClient.Get(
		conf.Redis.Prefix + key).Result(); err != redis.Nil {
		w := weights{}
		if err := json.Unmarshal([]byte(val), &w); err == nil {
			wb := w[key]
			r := rand.Intn(wb.TW) + 1
			for idx, cw := range wb.DW {
				sum = sum + cw.weight
				if r <= sum {
					wb.DW[idx].weight--
					retKey = wb.DW[idx].key
					wb.TW--
					if wb.TW == 0 {
						//free(wb.DW)
						wb.DW = make(kwA, 0)
						wb.ComputeWeights()
					}
					break
				}
			}
		} else {
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else { // we did not find it 404
		return c.JSON(http.StatusNotFound, nil)
	}

	return c.JSON(http.StatusOK, map[string]string{"key": retKey})
}

func deleteKey(c echo.Context) error {
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(http.StatusBadRequest, nil)
	}

	// Remove key from store
	err := rClient.Del("key").Err()
	if err != nil {
		return c.JSON(http.StatusNotFound, nil)
	}
	return c.JSON(http.StatusOK, "Deleted")
}

func main() {
	_init()
	runWeb()
}
