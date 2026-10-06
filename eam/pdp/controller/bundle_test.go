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
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestBase_NewBundle(t *testing.T) {
	t.Parallel()

	p1, _ := models.NewPolicyFromData("p1", "cedar", "", "", bytes.NewBufferString("allow=true"))
	p2, _ := models.NewPolicyFromData("p2", "cedar", "", "", bytes.NewBufferString("allow=true"))
	p3, _ := models.NewPolicyFromData("p3", "cedar", "", "", bytes.NewBufferString("allow=true"))

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

			ap1, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
			require.NoError(t, err)
			require.NotNil(t, ap1)

			_, _ = ap1.Create(p1, identity.NewPrincipal(identity.KindUser, "test"))
			_, _ = ap1.Create(p2, identity.NewPrincipal(identity.KindUser, "test"))
			_, _ = ap1.Create(p3, identity.NewPrincipal(identity.KindUser, "test"))

			ip1, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
			require.NoError(t, err)
			require.NotNil(t, ip1)

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

// TestBase_NewBundle_preservesOtherPIPSources verifies that applying a bundle never touches
// attributes/entities owned by other PIP sources (e.g. a local file store or a pull-config) -
// only attributes/entities the bundle itself provides are added or updated.
func TestBase_NewBundle_preservesOtherPIPSources(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	ap, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	ip, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
	require.NoError(t, err)

	// data owned by other PIP sources (e.g. a local file store / pull-config), not by any bundle.
	_, err = ip.AddAttribute(models.NewOriginalAttribute("a1", "original", "original", "xsd:string"))
	require.NoError(t, err)
	_, err = ip.AddAttribute(models.NewOriginalAttribute("a2", "original", "original", "xsd:string"))
	require.NoError(t, err)
	_, err = ip.AddEntity(models.NewEntity("user", "alice", models.NewAttributeSet()))
	require.NoError(t, err)
	_, err = ip.AddEntity(models.NewEntity("user", "bob", models.NewAttributeSet()))
	require.NoError(t, err)

	m := &Base{Ctx: ctx, Logger: logger, PAP: ap, PIP: ip, AuthMutex: &sync.RWMutex{}}

	bundle := &bundles.Bundle{
		Version: 1,
		Attributes: map[string]*attributes.Attribute{
			// "a2" also exists outside the bundle: the bundle's write should update it in place.
			"a2": {Key: "a2", Type: "xsd:string", Value: "from bundle"},
			"a4": {Key: "a4", Type: "xsd:string", Value: "hello mars"},
		},
		Entities: map[string]*attributes.Entity{
			"e4": {Type: "resource", Id: "book4"},
		},
	}

	_, err = m.NewBundle(bundle)
	require.NoError(t, err)

	a1, _, err := ip.GetAttribute("a1")
	require.NoError(t, err)
	assert.Equal(t, "original", a1.Value()) // untouched by the bundle.

	a2, _, err := ip.GetAttribute("a2")
	require.NoError(t, err)
	assert.Equal(t, "from bundle", a2.Value()) // updated by the bundle.

	a4, _, err := ip.GetAttribute("a4")
	require.NoError(t, err)
	assert.Equal(t, "hello mars", a4.Value()) // added by the bundle.

	alice, _, err := ip.GetEntity("user::alice")
	require.NoError(t, err)
	assert.NotNil(t, alice) // untouched by the bundle.

	bob, _, err := ip.GetEntity("user::bob")
	require.NoError(t, err)
	assert.NotNil(t, bob) // untouched by the bundle.

	book4, _, err := ip.GetEntity("resource::book4")
	require.NoError(t, err)
	assert.NotNil(t, book4) // added by the bundle.

	var attrCount, entityCount int

	ip.IterateAttributes(func(*models.Attribute) { attrCount++ })
	ip.IterateEntities(func(*models.Entity) { entityCount++ })
	assert.Equal(t, 3, attrCount)   // a1, a2, a4.
	assert.Equal(t, 3, entityCount) // alice, bob, book4.
}

// TestBase_NewBundle_prunesOnlyPreviousBundleData verifies that when a bundle stops carrying an attribute/entity it
// previously carried, that record is removed. Data owned by other PIP sources (never part of any bundle) survives
// across repeated bundle applications.
func TestBase_NewBundle_prunesOnlyPreviousBundleData(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))

	ap, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)

	ip, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
	require.NoError(t, err)

	// data owned by other PIP sources, never part of any bundle.
	_, err = ip.AddAttribute(models.NewAttribute("local", "kept"))
	require.NoError(t, err)
	_, err = ip.AddEntity(models.NewEntity("user", "alice", models.NewAttributeSet()))
	require.NoError(t, err)

	m := &Base{Ctx: ctx, Logger: logger, PAP: ap, PIP: ip, AuthMutex: &sync.RWMutex{}}

	_, err = m.NewBundle(&bundles.Bundle{
		Version: 1,
		Attributes: map[string]*attributes.Attribute{
			"b1": {Key: "b1", Type: "xsd:string", Value: "one"},
			"b2": {Key: "b2", Type: "xsd:string", Value: "two"},
		},
		Entities: map[string]*attributes.Entity{
			"r1": {Type: "resource", Id: "r1"},
			"r2": {Type: "resource", Id: "r2"},
		},
	})
	require.NoError(t, err)

	// second bundle drops "b1"/"r1" and adds "b3"/"r3".
	_, err = m.NewBundle(&bundles.Bundle{
		Version: 2,
		Attributes: map[string]*attributes.Attribute{
			"b2": {Key: "b2", Type: "xsd:string", Value: "two"},
			"b3": {Key: "b3", Type: "xsd:string", Value: "three"},
		},
		Entities: map[string]*attributes.Entity{
			"r2": {Type: "resource", Id: "r2"},
			"r3": {Type: "resource", Id: "r3"},
		},
	})
	require.NoError(t, err)

	b1, _, err := ip.GetAttribute("b1")
	require.NoError(t, err)
	assert.Nil(t, b1) // pruned: no longer in the latest bundle.

	b2, _, err := ip.GetAttribute("b2")
	require.NoError(t, err)
	assert.NotNil(t, b2) // kept: still in the latest bundle.

	b3, _, err := ip.GetAttribute("b3")
	require.NoError(t, err)
	assert.NotNil(t, b3) // added by the latest bundle.

	local, _, err := ip.GetAttribute("local")
	require.NoError(t, err)
	assert.NotNil(t, local) // never touched: not bundle-owned.

	r1, _, err := ip.GetEntity("resource::r1")
	require.NoError(t, err)
	assert.Nil(t, r1) // pruned: no longer in the latest bundle.

	r2, _, err := ip.GetEntity("resource::r2")
	require.NoError(t, err)
	assert.NotNil(t, r2) // kept: still in the latest bundle.

	r3, _, err := ip.GetEntity("resource::r3")
	require.NoError(t, err)
	assert.NotNil(t, r3) // added by the latest bundle.

	alice, _, err := ip.GetEntity("user::alice")
	require.NoError(t, err)
	assert.NotNil(t, alice) // never touched: not bundle-owned.
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

	wire, err := bundles.Decompress(bundles.CompressGZ, buf)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog2.NewDummyHandler(slog.LevelInfo))
	ap, err := pap.New(ctx, logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)
	require.NotNil(t, ap)

	ip, err := pip.New(ctx, logger, pip.WithKeyValueDB(memory.New(), ""))
	require.NoError(t, err)
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
	ap, err := pap.New(context.Background(), logger, pap.WithKeyValueDB(memory.New(), ""), pap.WithLanguage("cedar"))
	require.NoError(t, err)
	ip, err := pip.New(context.Background(), logger, pip.WithKeyValueDB(memory.New(), ""))
	require.NoError(t, err)

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
