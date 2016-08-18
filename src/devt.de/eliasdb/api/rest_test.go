/*
 * EliasDB
 *
 * Copyright 2016 Matthias Ladkau. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"devt.de/common/httputil"
)

const TESTPORT = ":9090"

var lastRes []string

type testEndpoint struct {
	*DefaultEndpointHandler
}

/*
handleSearchQuery handles a search query REST call.
*/
func (te *testEndpoint) HandleGET(w http.ResponseWriter, r *http.Request, resources []string) {
	lastRes = resources
	te.DefaultEndpointHandler.HandleGET(w, r, resources)
}

func (te *testEndpoint) SwaggerDefs(s map[string]interface{}) {
}

var testEndpointMap = map[string]RestEndpointInst{
	"/": func() RestEndpointHandler {
		return &testEndpoint{}
	},
}

func TestEndpointHandling(t *testing.T) {

	hs, wg := startServer()
	if hs == nil {
		return
	}

	queryURL := "http://localhost" + TESTPORT

	RegisterRestEndpoints(testEndpointMap)
	RegisterRestEndpoints(AboutEndpointMap)

	lastRes = nil

	if res := sendTestRequest(queryURL, "GET", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if lastRes != nil {
		t.Error("Unexpected lastRes:", lastRes)
	}

	lastRes = nil

	if res := sendTestRequest(queryURL+"/foo/bar", "GET", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if fmt.Sprint(lastRes) != "[foo bar]" {
		t.Error("Unexpected lastRes:", lastRes)
	}

	lastRes = nil

	if res := sendTestRequest(queryURL+"/foo/bar/", "GET", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if fmt.Sprint(lastRes) != "[foo bar]" {
		t.Error("Unexpected lastRes:", lastRes)
	}

	if res := sendTestRequest(queryURL, "POST", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if res := sendTestRequest(queryURL, "PUT", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if res := sendTestRequest(queryURL, "DELETE", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	if res := sendTestRequest(queryURL, "UPDATE", nil); res != "Method Not Allowed" {
		t.Error("Unexpected response:", res)
		return
	}

	// Test about endpoints

	if res := sendTestRequest(queryURL+"/db/about", "GET", nil); res != `
{
  "api_versions": [
    "v1"
  ],
  "product": "EliasDB",
  "version": "0.8"
}`[1:] {
		t.Error("Unexpected response:", res)
		return
	}

	if res := sendTestRequest(queryURL+"/db/swagger.json", "GET", nil); res != `
{
  "basePath": "/db",
  "definitions": {
    "Error": {
      "description": "A human readable error mesage.",
      "type": "string"
    }
  },
  "host": "localhost:9090",
  "info": {
    "description": "Query and modify the EliasDB datastore.",
    "title": "EliasDB API",
    "version": "1.0.0"
  },
  "paths": {
    "/about": {
      "get": {
        "description": "Returns available API versions, product name and product version.",
        "produces": [
          "text/plain",
          "application/json"
        ],
        "responses": {
          "200": {
            "description": "About info object",
            "schema": {
              "properties": {
                "api_versions": {
                  "description": "List of available API versions.",
                  "items": {
                    "description": "Available API version.",
                    "type": "string"
                  },
                  "type": "array"
                },
                "product": {
                  "description": "Product name of the REST API provider.",
                  "type": "string"
                },
                "version": {
                  "description": "Version of the REST API provider.",
                  "type": "string"
                }
              },
              "type": "object"
            }
          },
          "default": {
            "description": "Error response",
            "schema": {
              "$ref": "#/definitions/Error"
            }
          }
        },
        "summary": "Return information about the REST API provider."
      }
    }
  },
  "produces": [
    "application/json"
  ],
  "schemes": [
    "https"
  ],
  "swagger": "2.0"
}`[1:] {
		t.Error("Unexpected response:", res)
		return
	}

	stopServer(hs, wg)
}

/*
Send a request to a HTTP test server
*/
func sendTestRequest(url string, method string, content []byte) string {
	var req *http.Request
	var err error

	if content != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(content))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	bodyStr := strings.Trim(string(body), " \n")

	// Try json decoding first

	out := bytes.Buffer{}
	err = json.Indent(&out, []byte(bodyStr), "", "  ")
	if err == nil {
		return out.String()
	}

	// Just return the body

	return bodyStr
}

/*
Start a HTTP test server.
*/
func startServer() (*httputil.HTTPServer, *sync.WaitGroup) {
	hs := &httputil.HTTPServer{}

	var wg sync.WaitGroup
	wg.Add(1)

	go hs.RunHTTPServer(TESTPORT, &wg)

	wg.Wait()

	// Server is started

	if hs.LastError != nil {
		panic(hs.LastError)
	}

	return hs, &wg
}

/*
Stop a started HTTP test server.
*/
func stopServer(hs *httputil.HTTPServer, wg *sync.WaitGroup) {

	if hs.Running == true {

		wg.Add(1)

		// Server is shut down

		hs.Shutdown()

		wg.Wait()

	} else {

		panic("Server was not running as expected")
	}
}
