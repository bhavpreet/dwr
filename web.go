package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var (
	conf    dwrConfig
	rClient *redis.Client
)

func _init() {

	// Load json config file
	if fd, err := os.Open("config.json"); err == nil {
		_configJSON, _ = ioutil.ReadAll(fd)
		fd.Close()
	}

	if err := json.Unmarshal(_configJSON, &conf); err != nil {
		panic(err)
	}

	// env takes precedence
	redisNode := os.Getenv("REDIS_HOST")
	if len(redisNode) > 0 {
		conf.Redis.Address = redisNode
	}

	dbName := os.Getenv("REDIS_DB")
	if len(dbName) > 0 {
		if i, err := strconv.Atoi(dbName); err == nil {
			conf.Redis.DB = i
		}
	}

	rPass := os.Getenv("REDIS_PASS")
	if len(rPass) == 0 {
		conf.Redis.Password = rPass
	}

	// log.Printf(conf.Redis.Address)
	rClient = redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Address,
		Password: conf.Redis.Password, // no password set
		DB:       conf.Redis.DB,       // use default DB
	})
	log.Println("DWR initialized..")
	log.Println("With config ", conf)
}

func runWeb() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/:key", addKeyWeights)
	e.GET("/:key", getValue)
	e.DELETE("/:key", deleteKey)

	// Start server
	e.Logger.Fatal(e.Start(":80"))
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
	retString := "Created"

	// fmt.Println(c.Request().URL)
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(http.StatusBadRequest, nil)
	}

	wm := make(map[string]int)
	if err := c.Bind(&wm); err != nil {
		return c.JSON(http.StatusNotAcceptable, nil)
	}

	w := make(weights)
	w[key] = new(weightsBundle)
	w[key].UW = wm

	// Before we computeWeights; lets retrieve existing entry for the same
	// key if it exists and try to compare. If same, we can ignore
	if val, err := rClient.Get(
		conf.Redis.Prefix + key).Result(); err == nil {
		rW := make(weights)
		if err := json.Unmarshal([]byte(val), &rW); err != nil {
			if rW[key].UW != nil {
				if reflect.DeepEqual(w[key].UW, rW[key].UW) {
					return c.JSON(http.StatusCreated, "Same")
				} // else we compute
				retString = "Updated"
			} else if err == redis.Nil {
				// nothing, we compute
				retString = "Updated"
			} else if err != nil {
				panic(err)
			}
		}
	}

	// Compute weights
	w[key].DW = make(kwA, 0) // Create new DW for population
	w[key].ComputeWeights()

	// Write to store
	if jSr, err := json.Marshal(w); err == nil {
		if err := rClient.Set(conf.Redis.Prefix+key, jSr, 0).Err(); err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	} else {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, retString)
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
		conf.Redis.Prefix + key).Result(); err == nil {
		w := weights{}
		if err := json.Unmarshal([]byte(val), &w); err == nil {
			wb := w[key]
			r := rand.Intn(wb.TW) + 1
			for idx, cw := range wb.DW {
				sum = sum + cw.Weight
				if r <= sum {
					wb.DW[idx].Weight--
					retKey = wb.DW[idx].Key
					wb.TW--
					if wb.TW == 0 {
						//free(wb.DW)
						wb.DW = make(kwA, 0)
						wb.ComputeWeights()
					}
					break
				}
			}
			// We have to push the result back to store
			if jSr, err := json.Marshal(w); err == nil {
				if err := rClient.Set(
					conf.Redis.Prefix+key,
					jSr, 0).Err(); err != nil {
					return c.JSON(http.StatusInternalServerError, nil)
				}
			} else {
				return c.JSON(http.StatusInternalServerError, nil)
			}

		} else {
			return c.JSON(http.StatusInternalServerError, nil)
		}
	} else if err == redis.Nil { // we did not find it 404
		return c.JSON(http.StatusNotFound, nil)
	} else {
		panic(err)
		//return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"key": retKey})
}

func deleteKey(c echo.Context) error {
	key := c.Param("key")
	if len(key) == 0 {
		c.JSON(http.StatusBadRequest, nil)
	}

	// Remove key from store
	_, err := rClient.Del(conf.Redis.Prefix + key).Result()
	if err != nil {
		return c.JSON(http.StatusNotFound, nil)
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}

func main() {
	_init()
	runWeb()
}
