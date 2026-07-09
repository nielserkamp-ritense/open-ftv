// Package importer ingests ODRL-AP-NL documents (Turtle or JSON-LD) into the
// PAP policy store.
//
// Each policy in the document is stored under the "odrl" policy language as an
// envelope holding both the raw source and the parsed model. Referenced
// apnl:PolicyArtifacts with a known executable format (Rego/Cedar/OpenFGA) are
// downloaded, verified against their apnl:sha256, and stored under their own
// language key so an embedded PDP can execute them directly.
package importer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	pap2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pap"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pdp/odrl/model"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/io"
)

// MaxDocumentSize protects against unbounded downloads (bytes).
const MaxDocumentSize = 4 << 20 // 4 MiB

// Result reports what an import did.
type Result struct {
	Policies  []string `json:"policies"`           // stored PAP keys (odrl/...).
	Artifacts []string `json:"artifacts"`          // stored PAP keys for downloaded artifacts.
	Warnings  []string `json:"warnings,omitempty"` // non-fatal issues (per artifact).
}

// Importer ingests ODRL documents into a PAP.
type Importer struct {
	pap    pap2.PAP
	logger *slog.Logger
	client *http.Client
}

// New creates an Importer. client may be nil, in which case a default client
// with a 10s timeout is used.
func New(p pap2.PAP, logger *slog.Logger, client *http.Client) *Importer {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Importer{pap: p, logger: logger, client: client}
}

// Import parses and ingests a raw ODRL document ("text/turtle" or
// "application/ld+json"). The operation is idempotent on odrl:uid.
func (i *Importer) Import(ctx context.Context, source []byte, mimeType string) (*Result, error) {
	mimeType = normalizeMime(mimeType)

	doc, err := model.Parse(bytes.NewReader(source), mimeType)
	if err != nil {
		return nil, fmt.Errorf("odrl import: cannot parse document: %w", err)
	}

	warnings, err := doc.ValidateOpts(model.ValidateOptions{})
	if err != nil {
		return nil, fmt.Errorf("odrl import: document not valid ODRL-AP-NL: %w", err)
	}

	res := &Result{}
	for _, w := range warnings {
		i.logger.Warn("odrl import: " + w)
		res.Warnings = append(res.Warnings, w)
	}

	for _, pol := range doc.Policies {
		key, err2 := i.storePolicy(pol, doc, source, mimeType)
		if err2 != nil {
			return nil, err2
		}
		res.Policies = append(res.Policies, key)
	}

	i.importArtifacts(ctx, doc, res)
	return res, nil
}

// ImportURL fetches an ODRL document from the given URL and imports it.
func (i *Importer) ImportURL(ctx context.Context, url string) (*Result, error) {
	body, mimeType, err := i.fetch(ctx, url, fmt.Sprintf("%s;q=1,%s;q=0.5", mime.MimeTypeTurtle, mime.MimeTypeJSONLD))
	if err != nil {
		return nil, fmt.Errorf("odrl import: cannot fetch %s: %w", url, err)
	}
	return i.Import(ctx, body, mimeType)
}

// storePolicy stores one ODRL policy (and the document entities it needs)
// under the "odrl" language, keyed by a sanitized odrl:uid. Existing entries
// with the same uid are replaced (idempotent upsert).
func (i *Importer) storePolicy(pol *model.Policy, doc *model.Document, source []byte, mimeType string) (string, error) {
	sub := subDocument(pol, doc)
	env := model.NewEnvelope(pol.UID, mimeType, source, sub)

	data, err := env.Bytes()
	if err != nil {
		return "", err
	}

	id := SanitizeUID(pol.UID)
	stored, err2 := pap2.NewPolicyFromData(id, "odrl", "", pol.UID, bytes.NewReader(data))
	if err2 != nil {
		return "", err2
	}

	if prev, lastIndex, err3 := i.pap.Read("odrl", id); err3 == nil && prev != nil {
		if _, err4 := i.pap.Update(prev, lastIndex, stored); err4 != nil {
			return "", fmt.Errorf("odrl import: cannot update policy %s: %w", pol.UID, err4)
		}
		return stored.Key(), nil
	}

	if _, err3 := i.pap.Create(stored); err3 != nil {
		return "", fmt.Errorf("odrl import: cannot store policy %s: %w", pol.UID, err3)
	}
	return stored.Key(), nil
}

// importArtifacts downloads all referenced executable artifacts, verifies
// their sha256, and stores them under their own policy-language key. Failures
// are reported as warnings: the ODRL description itself was already stored.
func (i *Importer) importArtifacts(ctx context.Context, doc *model.Document, res *Result) {
	for _, a := range doc.Artifacts {
		key, err := i.importArtifact(ctx, a)
		switch {
		case err != nil:
			i.logger.Warn("odrl import: artifact skipped", "artifact", a.UID, "err", err)
			res.Warnings = append(res.Warnings, fmt.Sprintf("artifact %s: %v", a.UID, err))
		case key != "":
			res.Artifacts = append(res.Artifacts, key)
		}
	}
}

// importArtifact handles a single artifact. It returns ("", nil) when the
// artifact is not executable/downloadable (no downloadURL or unknown format).
func (i *Importer) importArtifact(ctx context.Context, a *model.Artifact) (string, error) {
	language, ok := artifactLanguage(a)
	if !ok || a.DownloadURL == "" {
		return "", nil
	}

	content, _, err := i.fetch(ctx, a.DownloadURL, "")
	if err != nil {
		return "", fmt.Errorf("download failed: %w", err)
	}

	if a.SHA256 == "" {
		return "", fmt.Errorf("no apnl:sha256 to verify against")
	}

	sum := sha256.Sum256(content)
	if got := hex.EncodeToString(sum[:]); !strings.EqualFold(got, a.SHA256) {
		return "", fmt.Errorf("sha256 mismatch: got %s, want %s", got, a.SHA256)
	}

	id := ArtifactID(a)
	stored, err2 := pap2.NewPolicyFromData(id, language, "", a.DownloadURL, bytes.NewReader(content))
	if err2 != nil {
		return "", err2
	}

	if prev, lastIndex, err3 := i.pap.Read(language, id); err3 == nil && prev != nil {
		if _, err4 := i.pap.Update(prev, lastIndex, stored); err4 != nil {
			return "", err4
		}
		return stored.Key(), nil
	}

	if _, err3 := i.pap.Create(stored); err3 != nil {
		return "", err3
	}
	return stored.Key(), nil
}

func (i *Importer) fetch(ctx context.Context, url, accept string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	if accept != "" {
		req.Header.Set("Accept", accept)
	}

	resp, err2 := i.client.Do(req)
	if err2 != nil {
		return nil, "", err2
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err3 := io.ReadAll(io.LimitReader(resp.Body, MaxDocumentSize+1))
	if err3 != nil {
		return nil, "", err3
	}
	if len(body) > MaxDocumentSize {
		return nil, "", fmt.Errorf("document exceeds %d bytes", MaxDocumentSize)
	}

	ct := resp.Header.Get("Content-Type")
	if idx := strings.IndexByte(ct, ';'); idx >= 0 {
		ct = ct[:idx]
	}
	return body, strings.TrimSpace(ct), nil
}

// subDocument narrows a parsed document to one policy plus all shared
// entities (artifacts, bundles, named constraints/duties) it may reference.
func subDocument(pol *model.Policy, doc *model.Document) *model.Document {
	return &model.Document{
		Policies:    []*model.Policy{pol},
		Artifacts:   doc.Artifacts,
		Bundles:     doc.Bundles,
		Constraints: doc.Constraints,
		Duties:      doc.Duties,
	}
}

// artifactLanguage maps an artifact to the PAP policy language it executes
// under, based on dct:format with the rdf:type as fallback.
func artifactLanguage(a *model.Artifact) (string, bool) {
	if l, ok := model.ArtifactLanguage(a.Format); ok {
		return l, true
	}
	return model.ArtifactLanguage(a.Type)
}

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9._~-]+`)

// SanitizeUID converts a policy uid (an IRI) into a single-segment PAP id.
// The mapping is deterministic, so re-importing the same uid overwrites the
// earlier version (idempotency).
func SanitizeUID(uid string) string {
	s := strings.TrimPrefix(uid, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "urn:")
	s = unsafeChars.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ArtifactID derives the PAP policy id for a downloaded artifact: the Rego
// entrypoint package path when available (e.g. "data.doelbinding.burgerzaken.allow"
// becomes "doelbinding/burgerzaken"), otherwise a sanitized uid.
func ArtifactID(a *model.Artifact) string {
	if strings.HasPrefix(a.Entrypoint, "data.") {
		segs := strings.Split(strings.TrimPrefix(a.Entrypoint, "data."), ".")
		if len(segs) >= 2 {
			segs = segs[:len(segs)-1] // drop the rule name, keep the package.
		}
		if len(segs) > 0 && segs[0] != "" {
			return strings.Join(segs, "/")
		}
	}
	return SanitizeUID(a.UID)
}

func normalizeMime(mimeType string) string {
	if idx := strings.IndexByte(mimeType, ';'); idx >= 0 {
		mimeType = mimeType[:idx]
	}
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))

	switch mimeType {
	case "", mime.MimeTypeTurtle, "application/x-turtle":
		return mime.MimeTypeTurtle
	case mime.MimeTypeJSONLD, "application/json+ld":
		return mime.MimeTypeJSONLD
	default:
		return mimeType
	}
}
