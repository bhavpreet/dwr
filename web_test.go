package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

var (
	c1 = []byte(`{
				"abc": 10,
				"bcd": 15,
				"cde": 25,
				"def": 50
			}`)

	c2 = []byte(`{
			"a":10,
			"b":20,
			"c":30
		  }`)
)

func createOrUpdate(e *echo.Echo, t *testing.T) {

	// Post
	req := httptest.NewRequest(echo.POST, "/", bytes.NewReader(c1))
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath("/:key")
	c.SetParamNames("key")
	c.SetParamValues("c1")

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	// Assertions
	if assert.NoError(t, addKeyWeights(c)) {
		// t.Log(rec.Code)
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func TestWeights(t *testing.T) {
	_init()
	e := echo.New()

	createOrUpdate(e, t)
	result := map[string]int{}

	jsonRes := map[string]string{}

	for i := 0; i < 100*5; i++ {
		// GET
		req := httptest.NewRequest(echo.GET, "/", bytes.NewReader(c1))
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		c.SetPath("/:key")
		c.SetParamNames("key")
		c.SetParamValues("c1")

		// try to simulate an update in between
		if i%20 == 0 {
			createOrUpdate(e, t)
		}

		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		// Assertions
		if assert.NoError(t, getValue(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
			if err := json.Unmarshal([]byte(rec.Body.String()),
				&jsonRes); err != nil {
				assert.Error(t, err, nil)
			}

			result[jsonRes["key"]]++
		}
	}

	assert.Equal(t, result["abc"], 10*5)
	assert.Equal(t, result["bcd"], 15*5)
	assert.Equal(t, result["cde"], 25*5)
	assert.Equal(t, result["def"], 50*5)
}
