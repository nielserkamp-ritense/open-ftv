package fiber

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/goccy/go-yaml"
	"github.com/gofiber/fiber/v2"

	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

func buildContent(req *fiber.Ctx, data any, modified *time.Time) error {
	if modified != nil {
		since := req.Get("If-Modified-Since")
		if t, err := time.Parse(http.TimeFormat, since); err == nil {
			if !modified.Truncate(time.Second).After(t) {
				return server.SendBasicResponse(req, fiber.StatusNotModified)
			}
		}
	}

	ct := fiber.MIMEApplicationJSON

	list := req.GetReqHeaders()[fiber.HeaderAccept]
	if len(list) > 0 && list[0] != "*/*" {
		ct = list[0]
	}

	var b []byte
	var err error

	switch strings.ToLower(ct) {
	case fiber.MIMEApplicationJSON, MimeJSON2:
		b, err = json.Marshal(data)
	case MimeYAML1, MimeYAML2:
		b, err = yaml.Marshal(data)
	case MimeCSV:
		b, err = marshalCSV(data)
	default:
		return server.SendMessageResponse(req, fiber.StatusBadRequest, "unsupported encoding: "+ct)
	}

	if err != nil {
		return server.SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}

	req.Set(fiber.HeaderContentType, ct)
	req.Set(fiber.HeaderContentLength, fmt.Sprintf("%d", len(b)))

	if modified != nil {
		req.Set(fiber.HeaderLastModified, modified.Format(http.TimeFormat))
	}

	return req.Send(b)
}

// MarshalCSV represents the interface to encode the underlying data to CSV format.
type MarshalCSV interface {
	MarshalCSV() ([]byte, error)
}

func marshalCSV(data any) ([]byte, error) {
	if f, ok := data.(MarshalCSV); ok {
		return f.MarshalCSV()
	}
	return nil, fmt.Errorf("marshalCSV: unsupported data type [%T]", data)
}
