package network

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func TestRequest_Prepare(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		method     string
		uri        string
		content    string
		parameters []*Parameter
		wantQuery  string
		wantURI    string
		wantBody   string
	}{
		{
			name:    "no parameters",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attributes",
			content: "application/json",
			wantURI: "http://localhost:9900/v1/attributes",
		},
		{
			name:    "single parameter - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
			},
			wantQuery: "key=hello",
			wantURI:   "http://localhost:9900/v1/attribute?key=hello",
		},
		{
			name:    "few parameters - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson&key=hello",
			wantURI:   "http://localhost:9900/v1/attribute?count=25&format=application%2Fjson&key=hello",
		},
		{
			name:    "few parameters & single attribute - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Attribute: "key"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson&key=%24invalid%24",
			wantURI:   "http://localhost:9900/v1/attribute?count=25&format=application%2Fjson&key=%24invalid%24",
		},
		{
			name:    "single parameter - path",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute/:key:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "path", Value: "hello", Type: "xsd:string"},
			},
			wantURI: "http://localhost:9900/v1/attribute/hello",
		},
		{
			name:    "few parameters - path",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedInt"},
			},
			wantURI: "http://localhost:9900/v1/entity/service/25",
		},
		{
			name:    "single parameter - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "body", Value: "hello", Type: "xsd:string"},
			},
			wantURI:  "http://localhost:9900/v1/attribute",
			wantBody: `{"key":"hello"}` + "\n",
		},
		{
			name:    "few parameters - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:unsignedLong"},
			},
			wantURI:  "http://localhost:9900/v1/entity",
			wantBody: "id: 25\ntype: service\n",
		},
		{
			name:    "few parameters & single attribute - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Attribute: "id"},
			},
			wantURI:  "http://localhost:9900/v1/entity",
			wantBody: "id: \"25\"\ntype: service\n",
		},
		{
			name:   "many parameters - body",
			method: "GET",
			uri:    "http://localhost:9900/v1/entity",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:anyType"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:byte"},
				{Name: "exact", In: "body", Value: 1, Type: "xsd:boolean"},
				{Name: "pct", In: "body", Value: 12.25, Type: "xsd:decimal"},
				{Name: "timeout", In: "body", Value: 10 * time.Second, Type: "xsd:duration"},
			},
			wantURI:  "http://localhost:9900/v1/entity",
			wantBody: `{"exact":true,"id":25,"pct":12.25,"timeout":"PT10S","type":"service"}` + "\n",
		},
		{
			name:    "mixed parameters",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedLong"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "hello", In: "body", Value: "world", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson",
			wantURI:   "http://localhost:9900/v1/entity/service/25?count=25&format=application%2Fjson",
			wantBody:  `{"hello":"world"}` + "\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Request{
				Name:        "request",
				Method:      tc.method,
				URI:         tc.uri,
				ContentType: tc.content,
				Parameters:  tc.parameters,
			}

			p.prepare()
			assert.Equal(t, tc.wantQuery, p.query)
			assert.Equal(t, tc.wantURI, p.uri)

			if tc.wantBody == "" {
				assert.Nil(t, p.body)
			} else {
				d, err := io.ReadAll(p.body)
				require.NoError(t, err)
				assert.Equal(t, tc.wantBody, string(d))
			}
		})
	}
}

func TestRequest_HTTPRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		method       string
		uri          string
		headers      map[string]string
		content      string
		timeout      time.Duration
		parameters   []*Parameter
		get          models.GetAttribute
		wantQuery    string
		wantURI      string
		wantBody     string
		wantErr2     bool
		wantURI2     string
		wantBody2    string
		wantHeaders2 map[string]string
	}{
		{
			name:     "bad url",
			method:   "GET",
			uri:      "\000\001\002",
			content:  "application/json",
			wantURI:  "\000\001\002",
			wantErr2: true,
		},
		{
			name:     "no parameters, no headers",
			method:   "GET",
			uri:      "http://localhost:9900/v1/attributes",
			content:  "application/json",
			wantURI:  "http://localhost:9900/v1/attributes",
			wantURI2: "http://localhost:9900/v1/attributes",
		},
		{
			name:         "no parameters, headers",
			method:       "GET",
			uri:          "http://localhost:9900/v1/attributes",
			headers:      map[string]string{"Hello": "world"},
			content:      "application/json",
			wantURI:      "http://localhost:9900/v1/attributes",
			wantURI2:     "http://localhost:9900/v1/attributes",
			wantHeaders2: map[string]string{"Hello": "world"},
		},
		{
			name:    "single parameter - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
			},
			wantQuery: "key=hello",
			wantURI:   "http://localhost:9900/v1/attribute?key=hello",
			wantURI2:  "http://localhost:9900/v1/attribute?key=hello",
		},
		{
			name:    "few parameters - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson&key=hello",
			wantURI:   "http://localhost:9900/v1/attribute?count=25&format=application%2Fjson&key=hello",
			wantURI2:  "http://localhost:9900/v1/attribute?count=25&format=application%2Fjson&key=hello",
		},
		{
			name:    "few parameters & single attribute - query",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			get:     getAttributeKey,
			parameters: []*Parameter{
				{Name: "key", In: "query", Attribute: "key"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
			},
			wantQuery: "count=25&key=%24invalid%24",
			wantURI:   "http://localhost:9900/v1/attribute?count=25&key=%24invalid%24",
			wantURI2:  "http://localhost:9900/v1/attribute?count=25&key=%24invalid%24",
		},
		{
			name:    "single parameter - path",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute/:key:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "path", Value: "hello", Type: "xsd:string"},
			},
			wantURI:  "http://localhost:9900/v1/attribute/hello",
			wantURI2: "http://localhost:9900/v1/attribute/hello",
		},
		{
			name:    "few parameters - path",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedInt"},
			},
			wantURI:  "http://localhost:9900/v1/entity/service/25",
			wantURI2: "http://localhost:9900/v1/entity/service/25",
		},
		{
			name:    "single parameter - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "body", Value: "hello", Type: "xsd:string"},
			},
			wantURI:      "http://localhost:9900/v1/attribute",
			wantBody:     `{"key":"hello"}` + "\n",
			wantURI2:     "http://localhost:9900/v1/attribute",
			wantHeaders2: map[string]string{"Content-Type": "application/json", "Content-Length": "16"},
		},
		{
			name:    "few parameters - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:unsignedLong"},
			},
			wantURI:      "http://localhost:9900/v1/entity",
			wantBody:     "id: 25\ntype: service\n",
			wantURI2:     "http://localhost:9900/v1/entity",
			wantHeaders2: map[string]string{"Content-Type": "application/yaml", "Content-Length": "21"},
		},
		{
			name:    "few parameters & single attribute - body",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity",
			content: "application/yaml",
			get:     getAttributeID,
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Attribute: "id"},
			},
			wantURI:      "http://localhost:9900/v1/entity",
			wantBody:     "id: \"25\"\ntype: service\n",
			wantURI2:     "http://localhost:9900/v1/entity",
			wantHeaders2: map[string]string{"Content-Type": "application/yaml", "Content-Length": "23"},
		},
		{
			name:   "many parameters - body",
			method: "GET",
			uri:    "http://localhost:9900/v1/entity",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:anyType"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:byte"},
				{Name: "exact", In: "body", Value: 1, Type: "xsd:boolean"},
				{Name: "pct", In: "body", Value: 12.25, Type: "xsd:decimal"},
				{Name: "timeout", In: "body", Value: 10 * time.Second, Type: "xsd:duration"},
			},
			wantURI:      "http://localhost:9900/v1/entity",
			wantBody:     `{"exact":true,"id":25,"pct":12.25,"timeout":"PT10S","type":"service"}` + "\n",
			wantURI2:     "http://localhost:9900/v1/entity",
			wantHeaders2: map[string]string{"Content-Type": "application/json", "Content-Length": "70"},
		},
		{
			name:    "mixed parameters, with headers",
			method:  "GET",
			uri:     "http://localhost:9900/v1/entity/:type:/:id:",
			headers: map[string]string{"API-KEY": "ab0cd1ef2gh3", "Accept-Encoding": "application/json"},
			content: "application/json",
			parameters: []*Parameter{
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedLong"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "hello", In: "body", Value: "world", Type: "xsd:string"},
			},
			wantQuery:    "count=25&format=application%2Fjson",
			wantURI:      "http://localhost:9900/v1/entity/service/25?count=25&format=application%2Fjson",
			wantBody:     `{"hello":"world"}` + "\n",
			wantURI2:     "http://localhost:9900/v1/entity/service/25?count=25&format=application%2Fjson",
			wantHeaders2: map[string]string{"Content-Type": "application/json", "Content-Length": "18", "API-KEY": "ab0cd1ef2gh3", "Accept-Encoding": "application/json"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := &Request{
				Name:        "request",
				Method:      tc.method,
				URI:         tc.uri,
				Headers:     tc.headers,
				ContentType: tc.content,
				Parameters:  tc.parameters,
			}

			p.prepare()
			assert.Equal(t, tc.wantQuery, p.query)
			assert.Equal(t, tc.wantURI, p.uri)

			if tc.wantBody == "" {
				assert.Nil(t, p.body)
			} else {
				require.NotNil(t, p.body)

				d, err := io.ReadAll(p.body)
				require.NoError(t, err)
				assert.Equal(t, tc.wantBody, string(d))
			}

			req, err := p.HTTPRequest(context.Background(), tc.get)
			if tc.wantErr2 {
				require.Error(t, err)
				require.Nil(t, req)
			} else {
				require.NoError(t, err)
				require.NotNil(t, req)

				assert.Equal(t, tc.method, req.Method)
				assert.Equal(t, tc.wantURI2, req.URL.String())

				for k := range tc.wantHeaders2 {
					list := req.Header.Values(k)
					require.Equal(t, 1, len(list))
					assert.Equal(t, tc.wantHeaders2[k], list[0])
				}

				if tc.wantBody2 != "" {
					require.NotNil(t, req.Body)

					d, err2 := io.ReadAll(req.Body)
					require.NoError(t, err2)
					assert.Equal(t, tc.wantBody2, string(d))
				}
			}
		})
	}
}

func getAttributeKey(_ string) any { return 25 }

func getAttributeID(_ string) any { return 123 }
