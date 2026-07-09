package server

import (
	"io"
	"strings"

	"github.com/gofiber/fiber/v2"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	srv "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
)

// GetPolicyByHash resolves a policy by the SHA-256 of its source content and returns
// the exact policy source recorded under that hash. It backs the content-addressable
// resolution path an ADL adl.core.policies reference {key: hash} needs: given the hash
// logged at decision time, a verifier retrieves the exact policy source that governed
// the decision - including older versions that have since been replaced (they remain
// retrievable under their own hash, with the documented retention implication that the
// index grows monotonically).
func (s *service) GetPolicyByHash(req *fiber.Ctx) error {
	hash := strings.ToLower(strings.TrimSpace(req.Params("hash")))
	if !isSHA256Hex(hash) {
		return srv.SendMessageResponse(req, fiber.StatusBadRequest, "hash must be a 64-character lowercase hex SHA-256")
	}

	pol, err := s.pap.ReadByHash(hash)
	if err != nil {
		return srv.SendMessageResponse(req, fiber.StatusNotFound, err.Error())
	}

	content, err := io.ReadAll(pol.Content())
	if err != nil {
		return srv.SendMessageResponse(req, fiber.StatusInternalServerError, err.Error())
	}

	// Integrity self-check: the stored content must still hash to the requested key.
	if pap2.HashContent(content) != hash {
		return srv.SendMessageResponse(req, fiber.StatusInternalServerError, "stored policy content no longer matches its hash")
	}

	req.Set("X-Policy-Id", pol.ID())
	req.Set("X-Policy-Language", pol.Language())
	req.Set("X-Policy-Hash", hash)
	req.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
	return req.Send(content)
}

// isSHA256Hex reports whether s is exactly 64 lowercase hex characters.
func isSHA256Hex(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}
