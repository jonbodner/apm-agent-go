// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

// This program is a test service for the W3C Distributed Tracing test harness:
//     https://github.com/w3c/trace-context/tree/master/test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"go.elastic.co/apm/module/apmhttp"
)

var (
	listenAddr = flag.String("listen", ":5000", "Address to listen on for test requests")
)

const (
	standardTraceparentHeader = "Traceparent"
	standardTracestateHeader  = "Tracestate"
)

var (
	tracestateKeyRegexp = regexp.MustCompile(`^[a-z](([a-z0-9_*/-]{0,255})|([a-z0-9_*/-]{0,240}@[a-z][a-z0-9_*/-]{0,13}))$`)

	// nblkchr is used for defining valid runes for tracestate values.
	nblkchr = &unicode.RangeTable{
		R16: []unicode.Range16{
			{0x21, 0x2B, 1},
			{0x2D, 0x3C, 1},
			{0x3E, 0x7E, 1},
		},
		LatinOffset: 3,
	}
)

func main() {
	flag.Parse()

	type TestCase struct {
		URL  string            `json:"url"`
		Args []json.RawMessage `json:"arguments,omitempty"`
	}

	client := http.DefaultClient
	client = apmhttp.WrapClient(client)

	var handler http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
		// We don't handle Tracestate in Elastic APM currently,
		// so we implement it in this test service just to pass
		// the test suite.
		var traceState traceState
		var traceStateOK bool
		if _, ok := req.Header[apmhttp.W3CTraceparentHeader]; ok {
			if values, ok := req.Header[standardTracestateHeader]; ok {
				traceStateOK = traceState.init(values...)
			}
		}

		var testCases []TestCase
		if err := json.NewDecoder(req.Body).Decode(&testCases); err != nil {
			log.Printf("decoding error: %s", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, tc := range testCases {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(tc.Args); err != nil {
				log.Printf("encoding error: %s", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			outReq, err := http.NewRequest("POST", tc.URL, &buf)
			if err != nil {
				log.Printf("error creating request: %s", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				continue
			}
			outReq = outReq.WithContext(req.Context())
			outReq.Header.Set("Content-Type", "application/json")
			if traceStateOK {
				outReq.Header.Set(standardTracestateHeader, traceState.String())
			}

			resp, err := client.Do(outReq)
			if err != nil {
				log.Printf("error sending request: %s", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				continue
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Printf("status: %s", resp.Status)
			}
		}
	}
	log.Printf("Starting Trace-Context test service, listening on %s", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, apmhttp.Wrap(handler)))
}

type traceState []traceStateItem

func (s *traceState) init(vs ...string) bool {
	recorded := make(map[string]bool)
	*s = (*s)[:0]
	for _, v := range vs {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		for _, field := range strings.Split(v, ",") {
			kv := strings.SplitN(strings.TrimSpace(field), "=", 2)
			if len(kv) != 2 {
				return false
			}
			item := traceStateItem{key: kv[0], value: kv[1]}
			if !tracestateKeyRegexp.MatchString(item.key) {
				return false
			}
			if !isValidValue(item.value) {
				return false
			}
			if recorded[item.key] {
				return false
			}
			recorded[item.key] = true
			*s = append(*s, item)
		}
	}
	return len(*s) <= 32
}

func isValidValue(s string) bool {
	runes := []rune(s)
	n := len(runes)
	if n == 0 || n > 256 {
		return false
	}
	if !unicode.In(runes[n-1], nblkchr) {
		return false
	}
	for _, r := range runes[:n-1] {
		if r != 0x20 && !unicode.In(r, nblkchr) {
			return false
		}
	}
	return true
}

func (s *traceState) String() string {
	fields := make([]string, len(*s))
	for i, item := range *s {
		fields[i] = fmt.Sprintf("%s=%s", item.key, item.value)
	}
	return strings.Join(fields, ",")
}

type traceStateItem struct {
	key   string
	value string
}
