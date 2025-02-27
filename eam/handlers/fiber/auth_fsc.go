package fiber

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/log/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/components/pdp"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/models"
	fiber2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/oas/fsc/auth"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/convert"
)

// AuthFSCVersion is the full semantic API version for the FSC authorization endpoint.
const AuthFSCVersion = "1.0.0"

// FSCAuthorizer represents the interface for handling FSC authorization requests.
type FSCAuthorizer interface {
	Authorize(req *fiber.Ctx) error
}

// NewAuthHandlerFSC instantiates a new FSC authorization handler.
func NewAuthHandlerFSC(logger *slog.Logger, authLogger authlog.Logger, controller pdp.Controller) FSCAuthorizer {
	return &authFSC{logger: logger, authLogger: authLogger, controller: controller}
}

// Authorize implements the FSCAuthorizer interface.
func (h *authFSC) Authorize(fc *fiber.Ctx) error {
	p := authProcess{
		logger:     h.logger,
		authLogger: h.authLogger,
		controller: h.controller,
		fc:         fc,
		status:     fiber.StatusInternalServerError,
		started:    time.Now(),
	}

	if p.logger.Enabled(nil, slog.LevelInfo) {
		defer p.log()
	}
	if p.authLogger != nil {
		defer p.authLog()
	}

	p.fc.Set(HeaderVersion, AuthFSCVersion)

	req := p.verifyRequestFSC()
	if p.err != nil {
		p.logger.Error("FSC authorization handler failed", "error", p.err)
		return fiber2.SendMessageResponse(fc, p.status, p.msg)
	}

	p.newAuthRequestFSC(req)
	return p.authorizeFSC()
}

func (p *authProcess) verifyRequestFSC() *auth.AuthorizationRequest {
	p.status = fiber.StatusBadRequest

	if req := p.fc.Request(); len(req.Header.ContentType()) == 0 {
		req.Header.SetContentType(fiber.MIMEApplicationJSON)
	}

	req := &auth.AuthorizationRequest{}
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if req.Input == nil {
		p.msg, p.err = "invalid data", errors.New("input must be filled")
		return nil
	}

	if req.Input.Method == "" {
		p.msg, p.err = "invalid method", errors.New("input.method must be filled")
		return nil
	}

	if req.Input.Path == "" {
		p.msg, p.err = "invalid path", errors.New("input.path must be filled")
		return nil
	}

	p.status = fiber.StatusOK
	return req
}

func (p *authProcess) newAuthRequestFSC(req *auth.AuthorizationRequest) {
	s := req.Input.Path
	if !strings.HasPrefix(s, "http") {
		s = fmt.Sprintf("https://%s", s)
	}
	if req.Input.Query != "" {
		s = fmt.Sprintf("%s?%s", s, req.Input.Query)
	}

	u, _ := url.ParseRequestURI(s)

	var d []byte
	if b := convert.OpaqueString(req.Input.Body); b != "" {
		d, _ = base64.StdEncoding.DecodeString(b)
	}

	uid, now := uuid.New(), time.Now().UTC()
	authReq := &models.Request{
		UID:         &uid,
		URL:         u,
		Method:      req.Input.Method,
		RequestTime: &now,
		Headers:     req.Input.Headers,
		Body:        d,
		PARC:        models.PARC{Context: models.NewAttributeSet()},
	}

	p.reqUID = uid.String()
	p.parc = p.controller.PEP().PARCFromRequest(authReq, p.controller.PIP())
}

func (p *authProcess) authorizeFSC() error {
	if p.resp, p.err = p.controller.Authorize(p.reqUID, p.parc); p.err != nil {
		p.msg = "FSC authorization process failed"
		return fiber2.SendMessageResponse(p.fc, p.status, p.msg)
	}

	allowed, msg := p.resp.Allowed, p.resp.Message
	if msg == "" {
		if allowed {
			msg = "ok"
		} else {
			msg = "not authorized"
		}
	}

	return p.fc.JSON(&auth.AuthorizationResponse{
		Result: &auth.AuthorizationResponseData{
			Allowed: &allowed,
			Status: &struct {
				Reason *string `json:"reason,omitempty"`
			}{Reason: &msg},
		},
	})
}

type authFSC struct {
	logger     *slog.Logger
	authLogger authlog.Logger
	controller pdp.Controller
}
