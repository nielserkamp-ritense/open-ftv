package network

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
)

func TestRequest_Prepare(t *testing.T) {
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
			uri:     "http://localhost:9000/v1/attributes",
			content: "application/json",
			wantURI: "http://localhost:9000/v1/attributes",
		},
		{
			name:    "single parameter - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
			},
			wantQuery: "key=hello",
			wantURI:   "http://localhost:9000/v1/attribute?key=hello",
		},
		{
			name:    "few parameters - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson&key=hello",
			wantURI:   "http://localhost:9000/v1/attribute?count=25&format=application%2Fjson&key=hello",
		},
		{
			name:    "few parameters & single attribute - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Attribute: "key"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
		},
		{
			name:    "single parameter - path",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute/:key:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "path", Value: "hello", Type: "xsd:string"},
			},
			wantURI: "http://localhost:9000/v1/attribute/hello",
		},
		{
			name:    "few parameters - path",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedInt"},
			},
			wantURI: "http://localhost:9000/v1/entity/service/25",
		},
		{
			name:    "single parameter - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "body", Value: "hello", Type: "xsd:string"},
			},
			wantURI:  "http://localhost:9000/v1/attribute",
			wantBody: `{"key":"hello"}`,
		},
		{
			name:    "few parameters - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:unsignedLong"},
			},
			wantURI:  "http://localhost:9000/v1/entity",
			wantBody: "id: 25\ntype: service\n",
		},
		{
			name:    "few parameters & single attribute - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Attribute: "id"},
			},
			wantURI: "http://localhost:9000/v1/entity",
		},
		{
			name:   "many parameters - body",
			method: "GET",
			uri:    "http://localhost:9000/v1/entity",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:anyType"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:byte"},
				{Name: "exact", In: "body", Value: 1, Type: "xsd:boolean"},
				{Name: "pct", In: "body", Value: 12.25, Type: "xsd:decimal"},
				{Name: "timeout", In: "body", Value: 10 * time.Second, Type: "xsd:duration"},
			},
			wantURI:  "http://localhost:9000/v1/entity",
			wantBody: `{"exact":true,"id":25,"pct":12.25,"timeout":"PT10S","type":"service"}`,
		},
		{
			name:    "mixed parameters",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedLong"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "hello", In: "body", Value: "world", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson",
			wantURI:   "http://localhost:9000/v1/entity/service/25?count=25&format=application%2Fjson",
			wantBody:  `{"hello":"world"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
	testCases := []struct {
		name       string
		method     string
		uri        string
		content    string
		parameters []*Parameter
		get        models.GetAttribute
		wantQuery  string
		wantURI    string
		wantBody   string
		wantURI2   string
		wantBody2  string
	}{
		{
			name:     "no parameters",
			method:   "GET",
			uri:      "http://localhost:9000/v1/attributes",
			content:  "application/json",
			wantURI:  "http://localhost:9000/v1/attributes",
			wantURI2: "http://localhost:9000/v1/attributes",
		},
		{
			name:    "single parameter - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
			},
			wantQuery: "key=hello",
			wantURI:   "http://localhost:9000/v1/attribute?key=hello",
			wantURI2:  "http://localhost:9000/v1/attribute?key=hello",
		},
		{
			name:    "few parameters - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "query", Value: "hello", Type: "xsd:string"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson&key=hello",
			wantURI:   "http://localhost:9000/v1/attribute?count=25&format=application%2Fjson&key=hello",
			wantURI2:  "http://localhost:9000/v1/attribute?count=25&format=application%2Fjson&key=hello",
		},
		{
			name:    "few parameters & single attribute - query",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			get:     &getAttributeKey{},
			parameters: []*Parameter{
				{Name: "key", In: "query", Attribute: "key"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
			},
			wantURI2: "http://localhost:9000/v1/attribute?count=25&key=25",
		},
		{
			name:    "single parameter - path",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute/:key:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "path", Value: "hello", Type: "xsd:string"},
			},
			wantURI:  "http://localhost:9000/v1/attribute/hello",
			wantURI2: "http://localhost:9000/v1/attribute/hello",
		},
		{
			name:    "few parameters - path",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedInt"},
			},
			wantURI:  "http://localhost:9000/v1/entity/service/25",
			wantURI2: "http://localhost:9000/v1/entity/service/25",
		},
		{
			name:    "single parameter - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/attribute",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "key", In: "body", Value: "hello", Type: "xsd:string"},
			},
			wantURI:  "http://localhost:9000/v1/attribute",
			wantBody: `{"key":"hello"}`,
			wantURI2: "http://localhost:9000/v1/attribute",
		},
		{
			name:    "few parameters - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity",
			content: "application/yaml",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:unsignedLong"},
			},
			wantURI:  "http://localhost:9000/v1/entity",
			wantBody: "id: 25\ntype: service\n",
			wantURI2: "http://localhost:9000/v1/entity",
		},
		{
			name:    "few parameters & single attribute - body",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity",
			content: "application/yaml",
			get:     &getAttributeID{},
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "body", Value: 25, Attribute: "id"},
			},
			wantURI:   "http://localhost:9000/v1/entity",
			wantURI2:  "http://localhost:9000/v1/entity",
			wantBody2: "id: 123\ntype: service\n",
		},
		{
			name:   "many parameters - body",
			method: "GET",
			uri:    "http://localhost:9000/v1/entity",
			parameters: []*Parameter{
				{Name: "type", In: "body", Value: "service", Type: "xsd:anyType"},
				{Name: "id", In: "body", Value: 25, Type: "xsd:byte"},
				{Name: "exact", In: "body", Value: 1, Type: "xsd:boolean"},
				{Name: "pct", In: "body", Value: 12.25, Type: "xsd:decimal"},
				{Name: "timeout", In: "body", Value: 10 * time.Second, Type: "xsd:duration"},
			},
			wantURI:  "http://localhost:9000/v1/entity",
			wantBody: `{"exact":true,"id":25,"pct":12.25,"timeout":"PT10S","type":"service"}`,
			wantURI2: "http://localhost:9000/v1/entity",
		},
		{
			name:    "mixed parameters",
			method:  "GET",
			uri:     "http://localhost:9000/v1/entity/:type:/:id:",
			content: "application/json",
			parameters: []*Parameter{
				{Name: "format", In: "query", Value: "application/json", Type: "xsd:string"},
				{Name: "type", In: "path", Value: "service", Type: "xsd:string"},
				{Name: "id", In: "path", Value: 25, Type: "xsd:unsignedLong"},
				{Name: "count", In: "query", Value: 25, Type: "xsd:int"},
				{Name: "hello", In: "body", Value: "world", Type: "xsd:string"},
			},
			wantQuery: "count=25&format=application%2Fjson",
			wantURI:   "http://localhost:9000/v1/entity/service/25?count=25&format=application%2Fjson",
			wantBody:  `{"hello":"world"}`,
			wantURI2:  "http://localhost:9000/v1/entity/service/25?count=25&format=application%2Fjson",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
				require.NotNil(t, p.body)

				d, err := io.ReadAll(p.body)
				require.NoError(t, err)
				assert.Equal(t, tc.wantBody, string(d))
			}

			req, err := p.HTTPRequest(context.Background(), tc.get)
			require.NoError(t, err)
			assert.Equal(t, tc.method, req.Method)
			assert.Equal(t, tc.wantURI2, req.URL.String())

			if tc.wantBody2 != "" {
				require.NotNil(t, req.Body)

				d, err2 := io.ReadAll(req.Body)
				require.NoError(t, err2)
				assert.Equal(t, tc.wantBody2, string(d))
			}
		})
	}
}

type getAttributeKey struct{}

func (t *getAttributeKey) GetAttribute(_ string) models.Attribute {
	return models.NewAttributeWithType("key", 25, "xsd:integer")
}

type getAttributeID struct{}

func (t *getAttributeID) GetAttribute(_ string) models.Attribute {
	return models.NewAttributeWithType("id", 123, "xsd:short")
}
