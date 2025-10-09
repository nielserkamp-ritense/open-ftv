package controller

import (
	"context"
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/pip"
	slog2 "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/slog"
)

func TestBase_Search(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		attrs   *models.AttributeSet
		uid     string
		req     *models.PARC
		allowed []*models.Relation
		wantErr bool
		want    []string
	}{
		{
			name: "bad search",
			uid:  "t1",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			wantErr: true,
		},
		{
			name: "principal - no list in PIP",
			uid:  "t2",
			req: &models.PARC{
				Principal: models.NewEntity("user", "", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			wantErr: true,
		},
		{
			name: "action - no list in PIP",
			uid:  "t3",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			wantErr: true,
		},
		{
			name: "resource - no list in PIP",
			uid:  "t4",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "", nil),
			},
			wantErr: true,
		},
		{
			name:  "principal - find one",
			attrs: models.NewAttributeSet(models.NewAttribute("subjects.user", "mickey,minnie,goofy,donald,dagobert,kwik,kwek,kwak")),
			uid:   "t5",
			req: &models.PARC{
				Principal: models.NewEntity("user", "", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			allowed: []*models.Relation{models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil))},
			want:    []string{"mickey"},
		},
		{
			name:  "principal - find a few",
			attrs: models.NewAttributeSet(models.NewAttribute("subjects.user", "mickey,minnie,goofy,donald,dagobert,kwik,kwek,kwak")),
			uid:   "t6",
			req: &models.PARC{
				Principal: models.NewEntity("user", "", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			allowed: []*models.Relation{
				models.NewRelation(models.NewEntity("user", "goofy", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil)),
				models.NewRelation(models.NewEntity("user", "minnie", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil)),
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil)),
			},
			want: []string{"mickey", "goofy", "minnie"},
		},
		{
			name:  "action - find one",
			attrs: models.NewAttributeSet(models.NewAttribute("actions.api", "HEAD,GET,PUT,POST,DELETE,OPTIONS")),
			uid:   "t7",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			allowed: []*models.Relation{models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil))},
			want:    []string{"POST"},
		},
		{
			name:  "action - find a few",
			attrs: models.NewAttributeSet(models.NewAttribute("actions.api", "HEAD,GET,PUT,POST,DELETE,OPTIONS")),
			uid:   "t8",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "", nil),
				Resource:  models.NewEntity("api", "brp", nil),
			},
			allowed: []*models.Relation{
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "HEAD", nil), models.NewEntity("api", "brp", nil)),
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "GET", nil), models.NewEntity("api", "brp", nil)),
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "OPTIONS", nil), models.NewEntity("api", "brp", nil)),
			},
			want: []string{"GET", "HEAD", "OPTIONS"},
		},
		{
			name:  "resource - find one",
			attrs: models.NewAttributeSet(models.NewAttribute("resources.api", "brp,kentekens,subsidies,vergunningen")),
			uid:   "t9",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "", nil),
			},
			allowed: []*models.Relation{models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "kentekens", nil))},
			want:    []string{"kentekens"},
		},
		{
			name:  "resource - find a few",
			attrs: models.NewAttributeSet(models.NewAttribute("resources.api", "brp,kentekens,subsidies,vergunningen")),
			uid:   "t10",
			req: &models.PARC{
				Principal: models.NewEntity("user", "mickey", nil),
				Action:    models.NewEntity("name", "POST", nil),
				Resource:  models.NewEntity("api", "", nil),
			},
			allowed: []*models.Relation{
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "brp", nil)),
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "subsidies", nil)),
				models.NewRelation(models.NewEntity("user", "mickey", nil), models.NewEntity("name", "POST", nil), models.NewEntity("api", "vergunningen", nil)),
			},
			want: []string{"brp", "subsidies", "vergunningen"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			h := slog2.NewDummyHandler(slog.LevelDebug)
			logger := slog.New(h)

			ip := pip.New(ctx, logger)
			require.NotNil(t, ip)
			if tc.attrs != nil {
				ip.MergeAttributes(tc.attrs)
			}

			b := &simpleEngine{Base: NewBase(WithContext(ctx), WithLogger(logger), WithPIP(ip)), allowed: tc.allowed}
			b.Self = b

			got, err := b.Search(tc.uid, tc.req)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)

				slices.Sort(tc.want)
				slices.Sort(got)
				require.EqualValues(t, tc.want, got)
			}
		})
	}
}

// poor man's ReBAC :)
type simpleEngine struct {
	Base
	allowed []*models.Relation
}

func (e *simpleEngine) Authorize(_ string, req *models.PARC) (*models.Response, error) {
	for i := range e.allowed {
		rel := e.allowed[i]
		if rel.Subject().Type() == req.Principal.Type() && rel.Subject().ID() == req.Principal.ID() &&
			rel.Predicate().ID() == req.Action.ID() &&
			rel.Object().Type() == req.Resource.Type() && rel.Object().ID() == req.Resource.ID() {
			return &models.Response{Allowed: true}, nil
		}
	}
	return &models.Response{Allowed: false}, nil
}

func (e *simpleEngine) Batch(uid string, req *models.Batch) ([]models.Response, error) {
	out := make([]models.Response, 0, len(req.Items))
	for i := range req.Items {
		resp, err := e.Authorize(uid, &req.Items[i])
		if err != nil {
			return nil, err
		}
		out = append(out, *resp)
	}
	return out, nil
}
