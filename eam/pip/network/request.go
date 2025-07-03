package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"k8s.io/client-go/transport"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/xsd"
)

// Request defines a request to retrieve external attributes, entities and/or relations.
//
// Method defines the http method to use for the request.
// URI defines the API endpoint to handle the request.
// Headers defines optional headers to send with the request.
// ContentType defines the content type of the request body (if any).
// Timeout defines the timeout before the request is cancelled.
//
// CAFile can be used to specify the certificate authority for checking the server certificate.
// CertFile and KeyFile can be used to specify the client-side TLS configuration.
// Each field should contain the path to the corresponding file.
//
// Insecure turns off server certificate validation. DO NOT USE FOR PRODUCTION!
//
// Parameters defines zero, one, or more parameters to supply in the request.
//
// Either Interval or Schedule should be present.
// If both are specified, Interval takes precedence.
// The Interval defines how often the request should be made; e.g. the interval between subsequent requests.
// The schedule can be used to define an arbitrary schedule to make the request.
// It should be formatted according to standard crontab layout; e.g. "* * * * *" ;
// Where:
// - the first parameter represents the minute(s).
// - the second parameter represents the hour(s).
// - the third parameter represents day(s) of the month.
// - the forth parameter represents the month(s).
// - the fifth parameter represents the day(s) of week.
//
// InitialInterval can be used to define an initial interval for the first request after the services has started.
//
// Mapping defines how to map the retrieved data to attributes, entities and/or relations.
type Request struct {
	Name            string            `json:"name" yaml:"name" toml:"name"`
	Description     string            `json:"description,omitempty" yaml:"description,omitempty" toml:"description,omitempty"`
	Method          string            `json:"method,omitempty" yaml:"method,omitempty" toml:"method,omitempty"`
	URI             string            `json:"uri" yaml:"uri" toml:"uri"`
	Headers         map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" toml:"headers,omitempty"`
	ContentType     string            `json:"contentType,omitempty" yaml:"contentType,omitempty" toml:"contentType,omitempty"`
	Timeout         time.Duration     `json:"timeout,omitempty" yaml:"timeout,omitempty" toml:"timeout,omitempty"`
	CAFile          string            `json:"tlsCA,omitempty" yaml:"tlsCA,omitempty" toml:"tlsCA,omitempty"`
	CertFile        string            `json:"tlsCert,omitempty" yaml:"tlsCert,omitempty" toml:"tlsCert,omitempty"`
	KeyFile         string            `json:"tlsKey,omitempty" yaml:"tlsKey,omitempty" toml:"tlsKey,omitempty"`
	Insecure        bool              `json:"insecure,omitempty" yaml:"insecure,omitempty" toml:"insecure,omitempty"`
	Parameters      []*Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty" toml:"parameters,omitempty"`
	Interval        time.Duration     `json:"interval,omitempty" yaml:"interval,omitempty" toml:"interval,omitempty"`
	InitialInterval time.Duration     `json:"initialInterval,omitempty" yaml:"initialInterval,omitempty" toml:"initialInterval,omitempty"`
	Schedule        string            `json:"schedule,omitempty" yaml:"schedule,omitempty" toml:"schedule,omitempty"`
	Mapping         *ResponseMapping  `json:"mapping,omitempty" yaml:"mapping,omitempty" toml:"mapping,omitempty"`
	// hidden fields.
	uri     string               // Fully encoded uri.
	query   string               // Fully encoded query.
	body    io.Reader            // Fully encoded body.
	bodyLen int                  // Length of fully encoded body.
	tls     *transport.TLSConfig // TLS configuration.
}

// HTTPRequest returns an HTTP request for retrieving the external data.
func (r *Request) HTTPRequest(ctx context.Context, get models.GetAttribute) (*http.Request, error) {
	query, static := r.prepareQuery(get)
	uri, _ := r.prepareURI(query, static, get)
	body, bodyLen, _ := r.prepareBody(get)

	httpReq, err := http.NewRequestWithContext(ctx, r.Method, uri, body)
	if err != nil {
		return nil, err
	}

	for key := range r.Headers {
		httpReq.Header.Set(key, r.Headers[key])
	}

	if body != nil {
		httpReq.Header.Set("Content-Type", r.ContentType)
		httpReq.Header.Set("Content-Length", strconv.Itoa(bodyLen))
	}

	return httpReq, nil
}

// TLSConfig returns the TLS client configuration for http requests.
func (r *Request) TLSConfig() *transport.TLSConfig {
	return r.tls
}

func (r *Request) prepareQuery(get models.GetAttribute) (string, bool) {
	if r.query != "" {
		return r.query, true
	}

	q, static := url.Values{}, true

	for i := range r.Parameters {
		if param := r.Parameters[i]; strings.EqualFold(param.In, "query") {
			v, s := param.ValueString(get)
			static = static && s
			q.Add(param.Name, v)
		}
	}

	return q.Encode(), static
}

func (r *Request) prepareURI(query string, static bool, get models.GetAttribute) (string, bool) {
	if r.uri != "" {
		return r.uri, true
	}

	uri := r.URI

	for i := range r.Parameters {
		if param := r.Parameters[i]; strings.EqualFold(param.In, "path") {
			find := fmt.Sprintf(":%s:", param.Name)
			v, s := param.ValueString(get)
			static = static && s
			uri = strings.Replace(uri, find, v, -1)
		}
	}

	if query != "" {
		return fmt.Sprintf("%s?%s", uri, query), static
	}
	return uri, static
}

func (r *Request) prepareBody(get models.GetAttribute) (io.Reader, int, bool) {
	if r.body != nil {
		return r.body, r.bodyLen, true
	}

	static := true
	m := make(map[string]any)

	for i := range r.Parameters {
		if param := r.Parameters[i]; strings.EqualFold(param.In, "body") {
			v, t, s := param.ValueAny(get)
			static = static && s

			switch t {
			case xsd.URIBoolean, xsd.PrefixBoolean:
				v = convert.AnyToBool(v)
			case xsd.URIInteger, xsd.URILong, xsd.URINonPos, xsd.URINeg, xsd.URIInt, xsd.URIShort, xsd.URIByte,
				xsd.PrefixInteger, xsd.PrefixLong, xsd.PrefixNonPos, xsd.PrefixNeg, xsd.PrefixInt, xsd.PrefixShort, xsd.PrefixByte:
				v = convert.AnyToInt64(v)
			case xsd.URINonNeg, xsd.URIPos, xsd.URIULong, xsd.URIUInt, xsd.URIUShort, xsd.URIUByte, xsd.URIDay, xsd.URIMonth, xsd.URIYear,
				xsd.PrefixNonNeg, xsd.PrefixPos, xsd.PrefixULong, xsd.PrefixUInt, xsd.PrefixYear, xsd.PrefixUShort, xsd.PrefixUByte, xsd.PrefixDay, xsd.PrefixMonth:
				v = convert.AnyToUint64(v)
			case xsd.URIFloat, xsd.URIDouble, xsd.URIDecimal, xsd.PrefixFloat, xsd.PrefixDouble, xsd.PrefixDecimal:
				v = convert.AnyToFloat64(v)
			case xsd.URIDuration, xsd.PrefixDuration:
				v, _ = xsd.ToString(v, param.Type)
			default:
				v = convert.AnyToString(v)
			}

			m[param.Name] = v
		}
	}

	if len(m) == 0 {
		return nil, 0, true
	}

	buf := &bytes.Buffer{}
	switch r.ContentType {
	case mime.MimeTypeYAML:
		_ = yaml.NewEncoder(buf).Encode(m)
	default:
		_ = json.NewEncoder(buf).Encode(m)
		r.ContentType = mime.MimeTypeJSON
	}

	return buf, buf.Len(), static
}

func (r *Request) prepare() {
	if r.Method == "" {
		r.Method = http.MethodGet
	}

	dummy := &dummyGetter{}

	query, static := r.prepareQuery(dummy)
	if static {
		r.query = query
	}

	// if there is a non-static query, the URI will also not be static.
	// so we can only have a static URI if there is no query, or if it is static.
	if query == "" || static {
		if uri, static2 := r.prepareURI(query, true, dummy); static2 {
			r.uri = uri
		}
	}

	if body, bodyLen, static2 := r.prepareBody(dummy); static2 {
		r.body = body
		r.bodyLen = bodyLen
	}

	if r.CAFile != "" || r.CertFile != "" || r.KeyFile != "" || r.Insecure {
		r.tls = &transport.TLSConfig{
			CAFile:   r.CAFile,
			CertFile: r.CertFile,
			KeyFile:  r.KeyFile,
			Insecure: r.Insecure,
		}
	}
}

// dummyGetter is a GetAttribute interface which always returns a dummy Attribute with value nil.
type dummyGetter struct{}

func (n *dummyGetter) GetAttribute(key string) models.Attribute { return models.NewAttribute(key, nil) }
