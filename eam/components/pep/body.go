package pep

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

func (c *collector) decodeBody() {
	if c.req.Body == nil {
		return
	}

	ct, ok := c.parc.Context.GetAttributeValue("content-type").(string)
	if ok && ct != "" {
		ct = strings.ToLower(convert.RemoveHeaderParameters(ct))
	} else {
		ct = convert.ContentSniffer(c.req.Body)
	}

	f, ok2 := parsers[ct]
	if !ok2 {
		c.logger.Error("unsupported content-type", "content-type", ct)
		return
	}

	if err := f(c.req.Body, c.parc.Context); err != nil {
		c.logger.Error("failed to parse body", "content-type", ct, "error", err)
	}
}

type bodyParser func(body []byte, a models.AttributeSet) error

var parsers = map[string]bodyParser{
	"text/xml":              parseXML,
	"application/xml":       parseXML,
	"application/soap+xml":  parseXML,
	"application/atom+xml":  parseXML,
	"application/rss+xml":   parseXML,
	"application/rdf+xml":   parseXML,
	"application/xhtml+xml": parseXML,
	"application/xslt+xml":  parseXML,
	"text/json":             parseJSON,
	"application/json":      parseJSON,
	"application/ld+json":   parseJSON,
	"application/geo+json":  parseJSON,
}

func parseJSON(body []byte, a models.AttributeSet) error {
	m := make(map[string]any)
	if err := json.NewDecoder(bytes.NewBuffer(body)).Decode(&m); err != nil && err != io.EOF {
		return err
	}

	a.AddAttribute("body", m)
	return nil
}

func parseXML(body []byte, a models.AttributeSet) error {
	nodes := make([]xmlNode, 0)
	if err := xml.Unmarshal(body, &nodes); err != nil {
		return err
	}

	nodes2 := make([]map[string]any, len(nodes))
	for i := range nodes {
		node := nodes[i]
		nodes2[i] = map[string]any{node.name(): node.toMap()}
	}
	a.AddAttribute("body", nodes2)

	return nil
}
