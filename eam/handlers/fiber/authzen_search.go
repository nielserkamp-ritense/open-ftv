package fiber

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	server "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/server/fiber"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
)

// SearchSubject implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchSubject(fc *fiber.Ctx) error {
	processHeaders(fc)

	if !h.hasSearchSubject {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	p, finish := initAuthProcess(fc, h.logger, nil, h.controller)
	defer finish()

	p.search = searchSubject

	req := p.verifySearchAuthZEN()
	if p.err != nil || req == nil {
		p.logger.Error("AuthZEN subject search request invalid", "request", req, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}
	p.authReq = req

	p.createSearchAuthZEN(&req.Subject, &req.Resource, &req.Action, req.Context)
	p.logger.Debug("AuthZEN subject search request", "request", p.parc)

	return p.searchAuthZEN()
}

// SearchAction implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchAction(fc *fiber.Ctx) error {
	processHeaders(fc)

	if !h.hasSearchAction {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	p, finish := initAuthProcess(fc, h.logger, nil, h.controller)
	defer finish()

	p.search = searchAction

	req := p.verifyActionSearchAuthZEN()
	if p.err != nil || req == nil {
		p.logger.Error("AuthZEN action search request invalid", "request", req, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}
	p.authReq = req

	p.createSearchAuthZEN(oas.EntityToSearch(&req.Subject), oas.EntityToSearch(&req.Resource), &oas.Action{}, req.Context)
	p.logger.Debug("AuthZEN action search request", "request", p.parc)

	return p.searchAuthZEN()
}

// SearchResource implements the AuthZENAuthorizer interface.
func (h *AuthZENAuthorizer) SearchResource(fc *fiber.Ctx) error {
	processHeaders(fc)

	if !h.hasSearchResource {
		return server.SendMessageResponse(fc, fiber.StatusNotImplemented, utils.StatusMessage(fiber.StatusNotImplemented))
	}

	p, finish := initAuthProcess(fc, h.logger, nil, h.controller)
	defer finish()

	p.search = searchResource

	req := p.verifySearchAuthZEN()
	if p.err != nil || req == nil {
		p.logger.Error("AuthZEN resource search request invalid", "request", req, "error", p.err)
		return server.SendMessageResponse(fc, p.status, p.msg)
	}
	p.authReq = req

	p.createSearchAuthZEN(&req.Subject, &req.Resource, &req.Action, req.Context)
	p.logger.Debug("AuthZEN resource search request", "request", p.parc)

	return p.searchAuthZEN()
}

func (p *authProcess) verifySearchAuthZEN() *oas.SearchRequest {
	p.initVerification()

	req := new(oas.SearchRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if req.Subject.Type == "" {
		p.msg, p.err = "invalid subject", errors.New("subject type must be filled")
		return nil
	}

	if p.search != searchSubject && req.Subject.Id == "" {
		p.msg, p.err = "invalid subject", errors.New("subject id must be filled")
		return nil
	}

	if req.Action.Name == "" {
		p.msg, p.err = "invalid action", errors.New("action name must be filled")
		return nil
	}

	if req.Resource.Type == "" {
		p.msg, p.err = "invalid resource", errors.New("resource type must be filled")
		return nil
	}

	if p.search != searchResource && req.Resource.Id == "" {
		p.msg, p.err = "invalid resource", errors.New("resource id must be filled")
		return nil
	}

	p.status = fiber.StatusOK
	p.offset = offsetFromToken(req.Page.Token)
	p.limit = req.Page.Limit
	return req
}

func (p *authProcess) verifyActionSearchAuthZEN() *oas.SearchActionRequest {
	p.initVerification()

	req := new(oas.SearchActionRequest)
	if p.err = p.fc.BodyParser(req); p.err != nil {
		p.msg = "invalid input data"
		return nil
	}

	if req.Subject.Type == "" || req.Subject.Id == "" {
		p.msg, p.err = "invalid subject", errors.New("subject type&id must be filled")
		return nil
	}

	if req.Resource.Type == "" || req.Resource.Id == "" {
		p.msg, p.err = "invalid resource", errors.New("resource type&id must be filled")
		return nil
	}

	p.status = fiber.StatusOK
	p.offset = offsetFromToken(req.Page.Token)
	p.limit = req.Page.Limit
	return req
}

func (p *authProcess) createSearchAuthZEN(subject, resource *oas.SearchEntity, action *oas.Action, ctx map[string]any) {
	p.parc = &models.PARC{
		Principal: models.NewEntity(subject.Type, subject.Id, models.NewAttributeSet(subject.Properties)),
		Action:    models.NewEntity(models.EntityTypeName, action.Name, models.NewAttributeSet(action.Properties)),
		Resource:  models.NewEntity(resource.Type, resource.Id, models.NewAttributeSet(resource.Properties)),
		Context:   models.NewAttributeSet(ctx),
	}
}

func (p *authProcess) searchAuthZEN() error {
	var list []string

	p.controller.GetPIP().ReportDynamicData(func() {
		list, p.err = p.controller.Search(p.reqUID, p.parc)
	})

	if p.err != nil {
		p.msg = fmt.Sprintf("AuthZEN %s search failed", p.search.String())
		return server.SendMessageResponse(p.fc, p.status, p.msg)
	}

	var page *oas.PageResponse
	list, page = p.processPagination(list)
	return p.processSearch(list, page)
}

func (p *authProcess) processPagination(list []string) ([]string, *oas.PageResponse) {
	page := &oas.PageResponse{}

	if p.limit == 0 {
		p.limit = 10000
	}

	page.Total = len(list)

	if p.offset > 0 {
		list = list[p.offset:]
	}

	if len(list) > p.limit {
		page.NextToken = offsetToToken(p.offset + p.limit)
		list = list[:p.limit]
	}

	page.Count = len(list)
	return list, page
}

func (p *authProcess) processSearch(list []string, page *oas.PageResponse) error {
	switch p.search {
	case searchSubject:
		results := make([]oas.SearchResult, len(list))
		for i := range list {
			results[i] = oas.SearchResult{Type: p.parc.Principal.Type(), Id: list[i]}
		}
		p.authResp = &oas.SearchResponse{Results: results, Page: *page}

	case searchAction:
		results := make([]oas.SearchActionResult, len(list))
		for i := range list {
			results[i] = oas.SearchActionResult{Name: list[i]}
		}
		p.authResp = &oas.SearchActionResponse{Results: results, Page: *page}

	case searchResource:
		results := make([]oas.SearchResult, len(list))
		for i := range list {
			results[i] = oas.SearchResult{Type: p.parc.Resource.Type(), Id: list[i]}
		}
		p.authResp = &oas.SearchResponse{Results: results, Page: *page}
	}

	return p.fc.JSON(p.authResp)
}

func offsetFromToken(token string) int {
	if token == "" {
		return 0
	}

	b, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return 0
	}

	dec := string(b)
	if strings.HasPrefix(dec, offsetPrefix) {
		dec = dec[len(offsetPrefix):]
	}

	out, _ := strconv.ParseInt(dec, 10, 64)
	return int(out)
}

func offsetToToken(offset int) string {
	token := fmt.Sprintf("%s%d", offsetPrefix, offset)
	return base64.StdEncoding.EncodeToString([]byte(token))
}

const offsetPrefix = "offset:"
