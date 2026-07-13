package fiber

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/authorization"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	bundles2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestNewBundleReceiverHandler(t *testing.T) {
	t.Parallel()

	p1, _ := models.NewPolicyFromData("p1", "cedar", "", "", bytes.NewBufferString("allow=true"))
	p2, _ := models.NewPolicyFromData("p2", "cedar", "", "", bytes.NewBufferString("allow=true"))
	p3, _ := models.NewPolicyFromData("p3", "cedar", "", "", bytes.NewBufferString("allow=true"))

	a1 := models.NewOriginalAttribute("a1", 123, "123", "xsd:integer")
	a2 := models.NewOriginalAttribute("a2", "hello world", "hello world", "xsd:string")
	a3 := models.NewOriginalAttribute("a3", true, "1", "xsd:boolean")
	setA := models.NewAttributeSet(a1, a2, a3)

	e1 := models.NewEntity("user", "alice", models.NewAttributeSet(a1))
	e2 := models.NewEntity("user", "bob", models.NewAttributeSet(a3, a2))
	setE := models.NewEntitySet(e2, e1)

	t.Run("bundle receiver", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		h := slog2.NewDummyHandler(slog.LevelInfo)
		logger := slog.New(h)

		ap1 := pap.New(ctx, logger, pap.WithLanguage("cedar"))
		require.NotNil(t, ap1)

		_, _ = ap1.Create(p1, "test")
		_, _ = ap1.Create(p2, "test")
		_, _ = ap1.Create(p3, "test")

		ip1 := pip.New(ctx, logger)
		require.NotNil(t, ip1)

		ip1.MergeAttributes(setA)
		ip1.MergeEntities(setE)

		m := &myBundleHandler{
			Base: controller.Base{
				Ctx:           ctx,
				Logger:        logger,
				PAP:           ap1,
				PIP:           ip1,
				AuthMutex:     &sync.RWMutex{},
				BundleVersion: 1,
			},
		}

		auth := authorization.New(
			authorization.WithLogger(logger),
			authorization.WithContext(ctx),
			authorization.NoAuth(),
		)

		handler := NewBundleReceiverHandler(logger, m, auth)

		srv := fiber.New()
		srv.Post("/v1/bundle", handler.PostBundle)

		bundle := &bundles.Bundle{
			Version: 2,
			Policies: map[string]*policies.Policy{
				"p1": {Data: "allow=false", Id: "p1", Language: "cedar", Metadata: policies.Metadata{Description: "bah"}},
				"p7": {Data: "allow=false", Id: "p7", Language: "cedar", Metadata: policies.Metadata{Description: "beh"}},
			},
			Attributes: map[string]*attributes.Attribute{
				"a2": {Key: "a2", Type: "xsd:integer", Value: 123456789},
				"a4": {Key: "a4", Type: "xsd:string", Value: "hello mars"},
				"a6": {Key: "a6", Type: "xsd:boolean", Value: true},
				"a8": {Key: "a8", Type: "xsd:integer", Value: -999},
			},
			Entities: map[string]*attributes.Entity{
				"e7": {Type: "resource", Id: "book7", Attributes: []attributes.Attribute{{Key: "a", Value: "d"}}},
			},
		}

		buf := &bytes.Buffer{}
		err := bundle.Compress(bundles.CompressBZ2, buf)
		require.NoError(t, err)

		req := httptest.NewRequest(fiber.MethodPost, "/v1/bundle", buf)
		req.Header.Add(fiber.HeaderContentEncoding, "bzip2")

		resp, err2 := srv.Test(req, 30000)
		require.NoError(t, err2)
		require.NotNil(t, resp)

		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		got := new(bundles2.BundleActivated)
		err = json.NewDecoder(resp.Body).Decode(got)
		require.NoError(t, err)
		assert.Equal(t, 1, got.PreviousVersion)
	})
}

type myBundleHandler struct {
	controller.Base
}

func (h *myBundleHandler) Batch(uid string, req *models.Batch) ([]models.Response, error) {
	// TODO implement me
	panic("implement me")
}

func (h *myBundleHandler) Authorize(string, *models.PARC) (*models.Response, error) {
	return nil, fmt.Errorf("not implemented")
}
