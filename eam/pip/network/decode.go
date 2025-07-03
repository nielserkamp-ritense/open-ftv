package network

import (
	"bytes"
	"fmt"
	"io"
	"unicode"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

func (r *runner) decodeResponse() {
	dec := r.req.Mapping
	if dec == nil {
		r.err = fmt.Errorf("no response body decoder defined")
		r.logger.Error("failed to process response body", "error", r.err)
		return
	}

	resp := r.httpResp

	var data []byte
	if data, r.err = io.ReadAll(resp.Body); r.err != nil {
		r.logger.Error("failed to read response body", "error", r.err)
		return
	}

	ct := resp.Header.Get(models.HeaderContentType)
	if ct == "" {
		ct = mime.SnifStream(bytes.NewReader(data))
	}

	switch ct {
	case mime.MimeTypeYAML:
		r.err = yaml.Unmarshal(data, &r.data)
	case mime.MimeTypeTOML:
		r.err = toml.Unmarshal(data, &r.data)
	case mime.MimeTypeTurtle, mime.MimeTypeJSONLD:
		r.err = r.decodeRDF(data, ct)
	default:
		r.err = r.decodeJSON(data)
	}

	if r.err != nil {
		r.logger.Error("failed to decode response body", "content-type", ct, "error", r.err)
		return
	}

	r.decodeData(dec, resp.StatusCode)
}

func (r *runner) decodeJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	for i := range data {
		if !unicode.IsSpace(rune(data[i])) {
			if data[i] == '[' {
				r.data = make([]any, 0)
			}
			break
		}
	}

	return json.Unmarshal(data, &r.data)
}

func (r *runner) decodeData(dec *ResponseMapping, status int) {
	for _, obj := range dec.Attributes {
		r.decodeAttribute(obj)
	}

	for _, obj := range dec.Entities {
		r.decodeEntity(obj)
	}

	for _, obj := range dec.Relations {
		r.decodeRelation(obj)
	}

	for _, obj := range dec.StatusCodes {
		if r.decodeStatus(obj, status) {
			break
		}
	}

	r.msg = msgOK
}
