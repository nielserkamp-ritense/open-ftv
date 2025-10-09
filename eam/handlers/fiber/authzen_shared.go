package fiber

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func (p *authProcess) initVerification() {
	p.status = fiber.StatusBadRequest

	if req := p.fc.Request(); len(req.Header.ContentType()) == 0 {
		req.Header.SetContentType(fiber.MIMEApplicationJSON)
	}

	p.reqUID = uuid.New().String()
}
