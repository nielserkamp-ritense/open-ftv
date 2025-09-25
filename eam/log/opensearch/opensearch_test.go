package opensearch

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities-no-ci/opensearch"
)

func TestNewOpenSearchAndLog(t *testing.T) {
	t.Run("new OpenSearch with log", func(t *testing.T) {
		newLogger = newMockOS
		defer func() {
			newLogger = opensearch.NewLogger
		}()

		l, err := NewOpenSearch("index", "user", "pswd", "https://localhost:9200")
		require.NoError(t, err)
		require.NotNil(t, l)

		l2, ok := l.(*os)
		require.True(t, ok)
		require.NotNil(t, l2)

		l3, ok2 := l2.sink.(*mockOS)
		require.True(t, ok2)
		require.NotNil(t, l3)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err = l3.CreateIndex(ctx, "index", 1, 1)
		require.NoError(t, err)

		err = l.Log(ctx, false, &AuthRecord{
			RvvaID:          "abc",
			Principal:       models.NewEntity("user", "alice", models.NewAttributeSet(models.NewAttribute("oin", "12345678901234567890"))),
			Action:          models.NewEntity("http", "GET", nil),
			Resource:        models.NewEntity("service", "brp-personen", models.NewAttributeSet(models.NewAttribute("oin", "98765432109876543210"))),
			Decision:        true,
			DecisionContext: models.NewAttributeSet(models.NewAttribute("policy", "*all*"), models.NewAttribute("version", "2025-02-05.11")),
		})
		require.NoError(t, err)

		assert.Equal(t, 1, len(l3.entries))
	})
}

func TestNewOpenSearch_Fail(t *testing.T) {
	t.Run("new OpenSearch", func(t *testing.T) {
		newLogger = opensearch.NewLogger
		l, err := NewOpenSearch("index", "user", "pswd", "http://localhost:12345")
		require.Error(t, err)
		require.Nil(t, l)
	})
}

func newMockOS(_, _ string, _ []string) (opensearch.Logger, error) {
	return &mockOS{indexes: make(map[string]struct{}), entries: make([]any, 0)}, nil
}

type mockOS struct {
	indexes map[string]struct{}
	entries []any
}

func (m *mockOS) CreateIndex(_ context.Context, name string, _, _ int) error {
	if _, ok := m.indexes[name]; ok {
		return fmt.Errorf("already exists: [%s]", name)
	}
	m.indexes[name] = struct{}{}
	return nil
}

func (m *mockOS) DeleteIndexes(_ context.Context, names ...string) error {
	for i := range names {
		name := names[i]
		if _, ok := m.indexes[name]; !ok {
			return fmt.Errorf("not exists: [%s]", name)
		}
		delete(m.indexes, name)
	}
	return nil
}

func (m *mockOS) Log(_ context.Context, _ bool, rec opensearch.LogRecord) error {
	if _, ok := m.indexes[rec.Index]; !ok {
		return fmt.Errorf("not exists: [%s]", rec.Index)
	}
	m.entries = append(m.entries, rec.Data)
	return nil
}

func (m *mockOS) LogBulk(ctx context.Context, wait bool, records ...opensearch.LogRecord) error {
	for i := range records {
		if err := m.Log(ctx, wait, records[i]); err != nil {
			return err
		}
	}
	return nil
}
