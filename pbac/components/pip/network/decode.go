package network

import (
	"bytes"
	"fmt"
	"io"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/pelletier/go-toml/v2"

	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/io"
)

func (r *runner) decodeResponse() {
	dec := r.req.Decoder
	if dec == nil {
		r.err = fmt.Errorf("no response decoder defined")
		r.logger.Error("failed to process http response body", "error", r.err)
		return
	}

	resp := r.httpResp

	var data []byte
	if data, r.err = io.ReadAll(resp.Body); r.err != nil {
		r.logger.Error("failed to read http response body", "error", r.err)
		return
	}

	ct := resp.Header.Get("Content-Type")
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
		r.err = json.Unmarshal(data, &r.data)
	}

	if r.err != nil {
		r.logger.Error("failed to decode http response body", "content-type", ct, "error", r.err)
		return
	}

	r.decodeData(dec)
}

func (r *runner) decodeData(dec *Response) {
	for _, obj := range dec.Attributes {
		r.decodeAttribute(obj)
	}

	for _, obj := range dec.Entities {
		r.decodeEntity(obj)
	}

	for _, obj := range dec.Relations {
		r.decodeRelation(obj)
	}

	r.msg = msgOK
}
