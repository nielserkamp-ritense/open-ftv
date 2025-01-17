package fiber

import (
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
)

func (s *service) defaultMiddleware() {
	if s.RecoverPanics {
		// recover from panics.
		s.svc.Use(recover2.New())
	}

	if s.CorsOrigins != "" || s.CorsHeaders != "" {
		// add CORS support so we can use the browser.
		s.svc.Use(cors.New(cors.Config{
			AllowOrigins: s.CorsOrigins,
			AllowHeaders: s.CorsHeaders,
		}))
	}

	if s.HighSecurity {
		// default headers for enhanced security.
		s.svc.Use(helmet.New())
	}
}
