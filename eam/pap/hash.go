package pap

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// hashPrefix is the reserved pseudo-language under which the content-addressable
// policy index is kept in the same KV store. A policy key has the shape
// "<language>/<id>"; an index entry has the shape "hash/<sha256>". Because no real
// policy language is ever named "hash", the two never collide and List filters the
// index out (see wrapper.List).
const hashPrefix = "hash" + PathSeparator

// HashContent returns the lowercase-hex SHA-256 of the given policy source bytes.
// This is the canonical content hash used both by the PAP content-addressable index
// and by a PDP engine that stamps DecisionContext.policyHash, so a logged
// adl.core.policies reference {key: hash} resolves back to the exact policy source
// that governed the decision.
func HashContent(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

// policyContentHash returns HashContent of a policy's source content.
func policyContentHash(p Policy) (string, error) {
	b, err := io.ReadAll(p.Content())
	if err != nil {
		return "", err
	}
	return HashContent(b), nil
}
