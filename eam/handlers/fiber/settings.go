package fiber

import (
	"errors"
	"log/slog"
	"regexp"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	auth "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/management"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/settings"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/settings"
)

// SettingsVersion is the full semantic API version for the settings endpoint.
const SettingsVersion = "1.0.0" // check against oas/settings/openapi.yaml!

// maxHeaderTitleLength is the maximum allowed length of the header title.
const maxHeaderTitleLength = 200

var (
	// settingsHexColor matches a 6-digit hex color code, e.g. #F7E8E8.
	settingsHexColor = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

	errHeaderTitleRequired  = checkIssue{code: "E08005", msg: "header title must be filled"}
	errHeaderTitleTooLong   = checkIssue{code: "E08010", msg: "header title too long (max 200 characters)"}
	errHeaderColorRequired  = checkIssue{code: "E08015", msg: "header color must be filled"}
	errHeaderColorInvalid   = checkIssue{code: "E08020", msg: "header color must be a hex color code, e.g. #F7E8E8"}
	errTitleColorRequired   = checkIssue{code: "E08025", msg: "title color must be filled"}
	errTitleColorInvalid    = checkIssue{code: "E08030", msg: "title color must be a hex color code, e.g. #CFCFCF"}
	errLogoMediaTypeInvalid = checkIssue{code: "E08035", msg: "logo media type must be one of image/png, image/jpeg, image/svg+xml, image/webp"}
)

// SettingsHandler implements the interface for handling requests about the manager ui settings.
type SettingsHandler struct {
	logger  *slog.Logger
	svc     *management.SettingService
	secured bool // an authorizer decides; the routes must carry the Identify middleware.
	principalResolver
}

// NewSettingsHandler instantiates a settings handler. Every operation is decided by the
// setting service through the authorizer; a nil authorizer permits everything.
func NewSettingsHandler(logger *slog.Logger, store settings.SettingsPersister, authorizer authorization.RequestAuthorizer, opts ...HandlerOption) *SettingsHandler {
	return &SettingsHandler{
		logger:            logger,
		svc:               management.NewSettingService(store, authorizer),
		secured:           authorizer != nil,
		principalResolver: newPrincipalResolver(logger, opts),
	}
}

// caller is the request-scoped Principal this request runs as.
func (h *SettingsHandler) caller(req *fiber.Ctx) *authorization.RequestPrincipal {
	return callerOf(req, h.secured, h.logger)
}

// GetSettings retrieves the manager ui settings.
func (h *SettingsHandler) GetSettings(req *fiber.Ctx) error {
	req.Set(HeaderVersion, SettingsVersion)

	resp, err := h.svc.Get(req.UserContext(), h.caller(req))
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, resp)
}

// PutSettings replaces the manager ui settings.
func (h *SettingsHandler) PutSettings(req *fiber.Ctx) error {
	req.Set(HeaderVersion, SettingsVersion)

	s, code, err := h.checkBody(req)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	resp, err := h.svc.Update(req.UserContext(), h.caller(req), s)
	if err != nil {
		return h.fail(req, err)
	}

	return h.respond(req, resp)
}

// fail answers a failed service call: a denial is 403, and anything else (including a
// permitted caller that could not be recorded) 500 without the cause.
func (h *SettingsHandler) fail(req *fiber.Ctx, err error) error {
	if errors.Is(err, authorization.ErrForbidden) {
		return auth.Forbidden(req, err, h.logger)
	}

	return internalError(req, h.logger, err)
}

// badRequest logs a warning and returns a 400 with the given validation error's code and message.
func (h *SettingsHandler) badRequest(req *fiber.Ctx, code string, err error) error {
	h.logger.Warn("settings request rejected", "path", req.Path(), "err", err, "status", fiber.StatusBadRequest)
	return server.SendProblemResponse(req, fiber.StatusBadRequest, code, err.Error())
}

// checkBody parses and validates the request body, returning the Code and error to report via
// badRequest on failure. A nil error means the returned Settings is valid.
func (h *SettingsHandler) checkBody(req *fiber.Ctx) (*oas.Settings, string, error) {
	var s oas.Settings
	if err := req.BodyParser(&s); err != nil {
		return nil, codeBadRequest, err
	}

	chk := newFieldChecker().
		checkHeaderTitle(s.HeaderTitle).
		checkHeaderColor(s.HeaderColor).
		checkTitleColor(s.TitleColor).
		checkLogo(&s)

	if chk.checkFailed() {
		return nil, chk.firstCode(), chk.error()
	}

	return &s, "", nil
}

func (f *fieldChecker) checkHeaderTitle(title string) *fieldChecker {
	switch {
	case title == "":
		f.addIssue(errHeaderTitleRequired)
	case utf8.RuneCountInString(title) > maxHeaderTitleLength:
		f.addIssue(errHeaderTitleTooLong)
	}

	return f
}

func (f *fieldChecker) checkHeaderColor(color string) *fieldChecker {
	switch {
	case color == "":
		f.addIssue(errHeaderColorRequired)
	case !settingsHexColor.MatchString(color):
		f.addIssue(errHeaderColorInvalid)
	}

	return f
}

func (f *fieldChecker) checkTitleColor(color string) *fieldChecker {
	switch {
	case color == "":
		f.addIssue(errTitleColorRequired)
	case !settingsHexColor.MatchString(color):
		f.addIssue(errTitleColorInvalid)
	}

	return f
}

func (f *fieldChecker) checkLogo(s *oas.Settings) *fieldChecker {
	switch {
	case len(s.Logo) == 0:
		s.LogoMediaType = ""
	case !s.LogoMediaType.Valid():
		f.addIssue(errLogoMediaTypeInvalid)
	}

	return f
}
