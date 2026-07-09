// Package odrl_geo implements a pdp.Controller that evaluates ODRL-Geo-NL
// policies directly (spec.md §7). It loads geo policies from the PAP under the
// "odrl" language (as imported by eam/pdp/odrl/importer), interprets the
// geo-specific terms the shared ODRL-AP-NL parser deliberately leaves out, and
// produces AuthZEN decisions with "ja, mits" obligations.
package odrl_geo

import (
	"io"
	"strings"
	"sync"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	pdp "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/controller"
)

// Version is the version of this ODRL-Geo-NL PDP engine.
const Version = "0.1.0"

// Name is the engine display name.
const Name = "ODRL-Geo-NL"

// language is the PAP policy language this engine consumes.
const language = "odrl"

// NewController instantiates a new ODRL-Geo-NL controller.
func NewController(options ...pdp.Option) pdp.Controller {
	options = append(options, pdp.WithNameVersion(Name, Version))

	c := &controller{
		Base:   pdp.NewBase(options...),
		docs:   map[string]*document{},
		hashes: map[string]string{},
		uids:   map[string][]string{},
	}

	if c.PAP() != nil {
		c.PAP().AddEventSink(c)
		c.PAP().LoadFiles()
		c.loadAll()
	}

	c.Logger().Info("pdp controller initialized", "controller", c.String())
	return c
}

type controller struct {
	pdp.Base
	mu   sync.RWMutex
	docs map[string]*document // keyed by PAP policy id.
	// hashes maps each loaded geoPolicy UID to the SHA-256 content hash of the PAP
	// envelope it was parsed from, so a decision can stamp DecisionContext.policyHash
	// with a value the PAP resolves back to the exact policy source via ReadByHash.
	hashes map[string]string
	// uids maps a PAP policy id to the geoPolicy UIDs it contributed, so a reload or
	// removal of that id can drop the corresponding hash entries without leaking
	// stale versions.
	uids map[string][]string
}

// loadAll (re)loads every "odrl" policy currently in the PAP.
func (c *controller) loadAll() {
	list, err := c.PAP().List(language)
	if err != nil {
		c.Logger().Error("failed to list odrl policies", "controller", c.String(), "error", err)
		return
	}
	for i := range list {
		_, id := pap.SplitPolicyKey(list[i].Key())
		c.loadPolicy(id)
	}
}

// loadPolicy reads and parses one policy from the PAP into the geo model.
func (c *controller) loadPolicy(id string) {
	f, _, err := c.PAP().Read(language, id)
	if err != nil {
		c.Logger().Error("failed to read odrl policy", "controller", c.String(), "policy-id", id, "error", err)
		return
	}
	data, err := io.ReadAll(f.Content())
	if err != nil {
		c.Logger().Error("failed to read odrl policy content", "controller", c.String(), "policy-id", id, "error", err)
		return
	}
	doc, err := loadDocument(data)
	if err != nil {
		c.Logger().Error("failed to parse odrl-geo policy", "controller", c.String(), "policy-id", id, "error", err)
		return
	}

	// Content hash of the exact envelope source: the same canonical hash the PAP
	// stores under its content-addressable index, so the audit reference resolves.
	hash := pap.HashContent(data)

	c.mu.Lock()
	c.dropHashes(id)
	c.docs[id] = doc
	uids := make([]string, 0, len(doc.Policies))
	for _, p := range doc.Policies {
		c.hashes[p.UID] = hash
		uids = append(uids, p.UID)
	}
	c.uids[id] = uids
	c.mu.Unlock()
	c.Logger().Info("odrl-geo policy loaded", "controller", c.String(), "policy-id", id, "policies", len(doc.Policies), "hash", hash)
}

// dropHashes removes the UID->hash entries a previous version of this PAP policy id
// contributed. The caller must hold c.mu.
func (c *controller) dropHashes(id string) {
	for _, uid := range c.uids[id] {
		delete(c.hashes, uid)
	}
	delete(c.uids, id)
}

// policyHash returns the content hash of the envelope a geoPolicy UID was parsed
// from, or "" when unknown (e.g. a default-deny with no governing policy).
func (c *controller) policyHash(uid string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hashes[uid]
}

// Handle implements the models.EventSink interface: it keeps the loaded geo
// policies in sync with the PAP.
func (c *controller) Handle(event models.EventType, key string) {
	lang, id := pap.SplitPolicyKey(key)
	if !strings.EqualFold(lang, language) {
		return
	}

	switch event {
	case models.PolicyAdded, models.PolicyReplaced:
		c.loadPolicy(id)
	case models.PolicyRemoved:
		c.mu.Lock()
		delete(c.docs, id)
		c.dropHashes(id)
		c.mu.Unlock()
		c.Logger().Info("odrl-geo policy removed", "controller", c.String(), "policy-id", id)
	}
}

// snapshot returns a copy of the currently loaded documents for evaluation.
func (c *controller) snapshot() []*document {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*document, 0, len(c.docs))
	for _, d := range c.docs {
		out = append(out, d)
	}
	return out
}
