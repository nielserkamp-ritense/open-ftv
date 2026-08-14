package fiber

import (
	"log/slog"
	"regexp"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
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
	logger     *slog.Logger
	store      settings.SettingsPersister
	authorizer authorization.Authorizer
}

// NewSettingsHandler instantiates a settings handler.
func NewSettingsHandler(logger *slog.Logger, store settings.SettingsPersister, authorizer authorization.Authorizer) *SettingsHandler {
	return &SettingsHandler{logger: logger, store: store, authorizer: authorizer}
}

// GetSettings retrieves the manager ui settings.
func (h *SettingsHandler) GetSettings(req *fiber.Ctx) error {
	req.Set(HeaderVersion, SettingsVersion)

	if _, err := h.authorize(req); err != nil {
		return err
	}

	resp, err2 := h.store.GetSettings(req.Context())
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(resp)
}

// PutSettings replaces the manager ui settings.
func (h *SettingsHandler) PutSettings(req *fiber.Ctx) error {
	req.Set(HeaderVersion, SettingsVersion)

	user, err := h.authorize(req)
	if err != nil {
		return err
	}

	s, code, err := h.checkBody(req)
	if err != nil {
		return h.badRequest(req, code, err)
	}

	resp, err2 := h.store.UpdateSettings(req.Context(), user, s)
	if err2 != nil {
		return h.error(req, fiber.StatusInternalServerError, err2)
	}

	return req.JSON(resp)
}

func (h *SettingsHandler) authorize(req *fiber.Ctx) (identity.Principal, error) {
	return authorizeRequest(h.authorizer, req, h.logger)
}

func (h *SettingsHandler) error(req *fiber.Ctx, status int, err error) error {
	h.logger.Error("request error", "path", req.Path(), "err", err, "status", status)
	return server.SendMessageResponse(req, status, err.Error())
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

	// Request body size is enforced by Fiber BodyLimit (MANAGER_MAX_BODY_SIZE).
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
