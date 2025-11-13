package bundles

import (
	"bytes"

	"github.com/google/uuid"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func newCedarPolicy(tag string) *models.Policy {
	uid := uuid.New().String()
	p, _ := models.NewPolicyFromData(uid, models.CEDAR.String(), "", "", bytes.NewBufferString(cedarDummy))
	return p.WithTags(tag).WithTitle("bootstrap policy").WithDescription("bootstrap policy to allow all requests (OpenFTV)")
}

func newRegoPolicy(tag string) *models.Policy {
	uid := uuid.New().String()
	p, _ := models.NewPolicyFromData(uid, models.REGO.String(), "", "", bytes.NewBufferString(regoDummy))
	return p.WithTags(tag).WithTitle("bootstrap policy").WithDescription("bootstrap policy to allow all requests (OpenFTV)")
}

func newCerbosPolicy(tag string) *models.Policy {
	uid := uuid.New().String()
	p, _ := models.NewPolicyFromData(uid, models.CERBOS.String(), "", "", bytes.NewBufferString(cerbosDummy))
	return p.WithTags(tag).WithTitle("bootstrap policy").WithDescription("bootstrap policy to allow all requests (OpenFTV)")
}

func newOpenFGAPolicy(tag string) *models.Policy {
	uid := uuid.New().String()
	p, _ := models.NewPolicyFromData(uid, models.OPENFGA.String(), "", "", bytes.NewBufferString(openfgaDummy))
	return p.WithTags(tag).WithTitle("bootstrap policy").WithDescription("bootstrap policy to allow all requests (OpenFTV)")
}

func newOpenFGARelations(tags []string) []*models.Relation {
	out := make([]*models.Relation, 0, len(openfgaUsers)*len(openfgaActions)*len(openfgaResources))

	for i := range openfgaUsers {
		for j := range openfgaActions {
			for k := range openfgaResources {
				out = append(out, models.NewRelation(
					models.NewEntity("user", openfgaUsers[i], nil),
					models.NewEntity("name", openfgaActions[j], nil),
					models.NewEntity("resource", openfgaResources[k], nil),
				).WithTags(tags...))
			}
		}
	}

	return out
}

const cedarDummy = `@comment("bootstrap policy to allow all requests (OpenFTV)")
permit (
    principal,
    action,
    resource
);
`

const regoDummy = `package authz

# bootstrap policy to allow all requests (OpenFTV)

default allow := true
`

const cerbosDummy = `---
# bootstrap policy to allow all requests (OpenFTV)
apiVersion: api.cerbos.dev/v1
resourcePolicy:
  resource: "*"
  version: "default"
  rules:
    - roles:
        - "*"
      actions:
        - "*"
      effect: EFFECT_ALLOW
`

const openfgaDummy = `model
  schema 1.1

type user

type resource
  relations
    define can_read: [user]
`

var (
	openfgaUsers     = []string{"alice", "bob"}
	openfgaActions   = []string{"can_read"}
	openfgaResources = []string{"BRP", "BRV"}
)
