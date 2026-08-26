package controller

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/bundles"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/identity"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/attributes"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/policies"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestBase_NewBundle(t *testing.T) {
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

	testCases := []struct {
		name   string
		bundle *bundles.Bundle
		lastV  uint64
	}{
		{
			name:   "empty",
			bundle: &bundles.Bundle{Version: 9},
			lastV:  3,
		},
		{
			name: "only policies",
			bundle: &bundles.Bundle{Version: 9, Policies: map[string]*policies.Policy{
				"p1": {Data: "allow=false", Id: "p1", Language: "cedar", Metadata: policies.Metadata{Description: "bah"}},
				"p4": {Data: "allow=false", Id: "p4", Language: "cedar", Metadata: policies.Metadata{Description: "boh"}},
				"p5": {Data: "allow=false", Id: "p5", Language: "cedar", Metadata: policies.Metadata{Description: "bih"}},
				"p6": {Data: "allow=false", Id: "p6", Language: "cedar", Metadata: policies.Metadata{Description: "buh"}},
				"p7": {Data: "allow=false", Id: "p7", Language: "cedar", Metadata: policies.Metadata{Description: "beh"}},
			}},
			lastV: 3,
		},
		{
			name: "only attributes",
			bundle: &bundles.Bundle{Version: 9, Attributes: map[string]*attributes.Attribute{
				"a2": {Key: "a2", Type: "xsd:integer", Value: 123456789},
				"a4": {Key: "a4", Type: "xsd:string", Value: "hello mars"},
				"a6": {Key: "a6", Type: "xsd:boolean", Value: true},
				"a8": {Key: "a8", Type: "xsd:integer", Value: -999},
				"a9": {Key: "a9", Type: "xsd:float", Value: 1234.56789},
			}},
			lastV: 8,
		},
		{
			name: "only entities",
			bundle: &bundles.Bundle{Version: 9, Entities: map[string]*attributes.Entity{
				"e1": {Type: "resource", Id: "book1", Attributes: []attributes.Attribute{{Key: "a", Value: "b"}}},
				"e4": {Type: "resource", Id: "book4", Attributes: []attributes.Attribute{{Key: "x", Value: "y"}}},
				"e5": {Type: "resource", Id: "book5", Attributes: []attributes.Attribute{{Key: "b", Value: "c"}}},
				"e7": {Type: "resource", Id: "book7", Attributes: []attributes.Attribute{{Key: "a", Value: "d"}}},
			}},
			lastV: 8,
		},
		{
			name: "all",
			bundle: &bundles.Bundle{
				Version: 9,
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
			},
			lastV: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelInfo)
			logger := slog.New(h)

			ap1 := pap.New(ctx, logger, pap.WithLanguage("cedar"))
			require.NotNil(t, ap1)

			_, _ = ap1.Create(p1, identity.NewPrincipal(identity.KindUser, "test"))
			_, _ = ap1.Create(p2, identity.NewPrincipal(identity.KindUser, "test"))
			_, _ = ap1.Create(p3, identity.NewPrincipal(identity.KindUser, "test"))

			ip1 := pip.New(ctx, logger)
			require.NotNil(t, ip1)

			ip1.MergeAttributes(setA)
			ip1.MergeEntities(setE)

			m := &Base{
				Ctx:           ctx,
				Logger:        logger,
				PAP:           ap1,
				PIP:           ip1,
				AuthMutex:     &sync.RWMutex{},
				BundleVersion: tc.lastV,
			}

			got, err := m.NewBundle(tc.bundle)
			require.NoError(t, err)
			assert.Equal(t, tc.lastV, got)

			var count int
			ap1.Iterate(func(policy *models.Policy) {
				count++

				var ok bool
				for _, other := range tc.bundle.Policies {
					if other.Language == policy.Language() && other.Id == policy.ID() {
						ok = true
						break
					}
				}
				assert.True(t, ok)
			})
			assert.Equal(t, len(tc.bundle.Policies), count)

			count = 0
			ip1.IterateAttributes(func(attr *models.Attribute) {
				count++

				var ok bool
				for _, other := range tc.bundle.Attributes {
					if other.Key == attr.Key() {
						ok = true
						break
					}
				}
				assert.True(t, ok)
			})
			assert.Equal(t, len(tc.bundle.Attributes), count)

			count = 0
			ip1.IterateEntities(func(entity *models.Entity) {
				count++

				var ok bool
				for _, other := range tc.bundle.Entities {
					if other.Type == entity.Type() && other.Id == entity.ID() {
						ok = true
						break
					}
				}
				assert.True(t, ok)
			})
			assert.Equal(t, len(tc.bundle.Entities), count)
		})
	}
}

func TestBase_NewBundle_wireBundleFromAddPolicy(t *testing.T) {
	t.Parallel()

	p, err := models.NewPolicyFromData("generic", "cedar", "", "", bytes.NewBufferString(`permit (
    principal,
    action,
    resource is service
);`))
	require.NoError(t, err)
	p.WithTags("pdp1")

	src := bundles.NewBundle(2, "cedar", "pdp1")
	require.True(t, src.AddPolicy(p))

	buf := &bytes.Buffer{}
	require.NoError(t, src.Compress(bundles.CompressGZ, buf))

	wire, err := bundles.BundleFromAPI(buf, map[string][]string{"Content-Encoding": {"gzip"}})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	ap := pap.New(ctx, logger, pap.WithLanguage("cedar"))
	require.NotNil(t, ap)

	ip := pip.New(ctx, logger)
	require.NotNil(t, ip)

	m := &Base{
		Ctx:       ctx,
		Logger:    logger,
		PAP:       ap,
		PIP:       ip,
		AuthMutex: &sync.RWMutex{},
	}

	_, err = m.NewBundle(wire)
	require.NoError(t, err)

	var count int
	ap.Iterate(func(pol *models.Policy) {
		count++
		assert.Equal(t, "cedar", pol.Language())
		assert.Equal(t, "cedar/generic", pol.Key())
	})
	assert.Equal(t, 1, count)
}

func TestBase_processPolicies_emptyLanguage(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	ap := pap.New(context.Background(), logger)
	ip := pip.New(context.Background(), logger)

	m := &Base{
		PAP:       ap,
		PIP:       ip,
		Logger:    logger,
		AuthMutex: &sync.RWMutex{},
	}

	bundle := &bundles.Bundle{
		Version:  2,
		Language: "cedar",
		Policies: map[string]*policies.Policy{
			"p1": {Data: cedarDummy, Id: "p1"},
		},
	}

	require.NoError(t, m.processPolicies(bundle))

	got, _, err := ap.Read("p1")
	require.NoError(t, err)
	assert.Equal(t, "cedar", got.Language())
	assert.Equal(t, "cedar/p1", got.Key())
}

const cedarDummy = `@comment("test")
permit (
    principal,
    action,
    resource
);
`
