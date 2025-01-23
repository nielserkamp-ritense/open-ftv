package server

import "github.com/gofiber/fiber/v2"

// ledenlijst returns a fictive ledenlijst.
func ledenlijst(req *fiber.Ctx) error {
	req.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return req.Send(data)
}

var data = []byte(`[
 {"id":"dfa58ae2-0dd4-471d-8863-08147944e641","oin":"01726477373371538205","attributes":{"name":"FSC controller","isMember":true,"maturity":4}},
 {"id":"f8d73ae9-f3a7-4521-a6c0-28422b056cbb","oin":"01726469521943092351","attributes":{"name":"RDW","isMember":true,"maturity":3}},
 {"id":"86af57a8-c9d6-4620-8511-e81498ff2df7","oin":"01726469365987449994","attributes":{"name":"RViG","isMember":true,"maturity":5}}
]`)
