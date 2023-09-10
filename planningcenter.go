package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Make an API request to Planning Center.
func NewPCRequest(uri string) (*http.Request, error) {
	url := uri
	// If request URI doesn't include full URL, prepend the PC API URL.
	if !strings.HasPrefix(url, "http") {
		url = "https://api.planningcenteronline.com" + uri
	}
	// Make the request.
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Append the basic authentication from the configuration.
	auth := app.config.PlanningCenter.AppID + ":" + app.config.PlanningCenter.Secret
	authString := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+authString)

	// Return the request made.
	return req, nil
}

// Planning center meta data/information about request.
type PCMeta struct {
	TotalCount uint64 `json:"total_count"`
	Count      uint64 `json:"count"`

	Prev struct {
		Offset uint64 `json:"offset"`
	} `json:"prev"`
	Next struct {
		Offset uint64 `json:"offset"`
	} `json:"next"`

	CanOrderBy []string `json:"can_order_by"`
	CanQueryBy []string `json:"can_query_by"`
	CanInclude []string `json:"can_include"`

	Parent struct {
		Id   string `json:"id"`
		Type string `json:"type"`
	} `json:"parent"`
}

// A dictionary for planning center response parsing.
type PCDict map[string]interface{}

// Common response error structure.
type PCError struct {
	Status string `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Basic PC response structure.
type PCResponse struct {
	Links struct {
		Self string `json:"self"`
		Prev string `json:"prev"`
		Next string `json:"next"`
	} `json:"links"`
	Data     []PCDict      `json:"data"`
	Included []interface{} `json:"included"`
	Meta     PCMeta        `json:"meta"`
	Errors   []PCError     `json:"errors"`
}

// Parse a planning center reponse body.
func PCParseResponse(body io.Reader) (*PCResponse, error) {
	// Decode JSON response.
	res := new(PCResponse)
	err := json.NewDecoder(body).Decode(res)
	if err != nil {
		return nil, err
	}
	// If an error was provided from the API, return it.
	if len(res.Errors) != 0 {
		return nil, fmt.Errorf(res.Errors[0].Detail)
	}
	// We expect result to be provided on a valid response.
	if res.Data == nil {
		return nil, fmt.Errorf("no data in response")
	}
	// A valid response was decoded, return it.
	return res, nil
}

// Query Planning Center API and get data from all pages.
func PCGetAll(uri string) ([]PCDict, error) {
	// The data array to store all found data.
	var data []PCDict

	// Set the first URL to the requested URL.
	url := uri
	// Make requests until the last page was loaded.
	for {
		// Make the request.
		req, err := NewPCRequest(url)
		if err != nil {
			return nil, err
		}

		// Perform the request.
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		// Close body when done.
		defer res.Body.Close()

		// Parse the response.
		response, err := PCParseResponse(res.Body)
		if err != nil {
			return nil, err
		}

		// Add data from response to global data array.
		data = append(data, response.Data...)

		// If no next link provided, stop here.
		if response.Links.Next == "" {
			break
		}
		// If next link provided, set it for the next request.
		url = response.Links.Next
	}
	// Return all found data.
	return data, nil
}

// Below are a bunch of helper functions.
// I would recommend using a tool like Insomnia to test API requests,
// then you will know what the data structure is like for an API request.
// Planning center does have some ok documentation available:
// https://developer.planning.center/docs/#/overview

// Get a string from a dictionary.
func (p PCDict) GetString(key string) string {
	s, ok := p[key].(string)
	if !ok {
		return ""
	}
	return s
}

// Get a bool from a dictionary.
func (p PCDict) GetBool(key string) bool {
	b, ok := p[key].(bool)
	if !ok {
		return false
	}
	return b
}

// Get an unsigned int from dictionary.
func (p PCDict) GetUint64(key string) uint64 {
	s, ok := p[key].(string)
	var i uint64
	// Try parsing a string if its ok.
	if ok {
		i, _ = strconv.ParseUint(s, 10, 64)
	} else {
		// Otherwise, try converting to an integer.
		i, ok = p[key].(uint64)
		if !ok {
			return 0
		}
	}
	return i
}

// Get a dictionary from a dictionary.
func (p PCDict) GetDict(key string) PCDict {
	d, ok := p[key].(map[string]interface{})
	if !ok {
		return make(map[string]interface{})
	}
	return d
}

// Standard date layouts.
const (
	PCDateTimeLayout = "2006-01-02T15:04:05Z"
	PCDateLayout     = "2006-01-02"
)

// Get a date from a dictionary.
func (p PCDict) GetDate(key string) time.Time {
	var t time.Time
	var err error
	s, ok := p[key].(string)
	if ok {
		// Try parsing with the time layout first.
		t, err = time.Parse(PCDateTimeLayout, s)
		if err != nil {
			// If that fialed, try using the date layout.
			t, _ = time.Parse(PCDateLayout, s)
		}
	}
	return t
}
