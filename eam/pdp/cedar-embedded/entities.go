package cedar_embedded

import (
	"strings"

	"github.com/cedar-policy/cedar-go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
)

func entityToCedar(in *models.Entity) (*cedar.Entity, error) {
	if in == nil {
		return nil, nil
	}

	parents := make([]cedar.EntityUID, len(in.Parents()))
	for i := range in.Parents() {
		parents[i] = uidToCedar(in.Parents()[i])
	}

	s := make(cedar.RecordMap)
	var err error
	in.Attributes().IterateAttributes(func(attr *models.Attribute) {
		if err != nil {
			return
		}
		s[cedar.String(attr.Key())], err = anyToValue(attr.Value())
	})

	return &cedar.Entity{
		UID:        uidToCedar(in.UID()),
		Parents:    cedar.NewEntityUIDSet(parents...),
		Attributes: cedar.NewRecord(s),
	}, err
}

func uidToCedar(uid string) cedar.EntityUID {
	parts := strings.Split(uid, "::")
	return cedar.NewEntityUID(cedar.EntityType(parts[0]), cedar.String(parts[1]))
}
