package bundles

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/valkeyrie/memory"
)

func TestNewDeployer(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		basePath string
	}{
		{name: "no prefix"},
		{name: "prefix no separator", basePath: "prefix"},
		{name: "prefix with separator", basePath: "prefix2/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			s := memory.New()
			require.NotNil(t, s)

			d := NewKeyValueDB(s, tc.basePath)
			require.NotNil(t, d)
			assert.Equal(t, s, d.client)
			assert.Equal(t, convert.ForceSuffix(tc.basePath, "/"), d.basePath)
		})
	}
}

func TestDeployer_Generate(t *testing.T) {
	t.Parallel()

	t.Run("generate", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Generate(ctx, "v1", "hello world", "")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, Creating, d2.Status())

		d3, err3 := d.Generate(ctx, "v2", "next one", "*SYSTEM*")
		require.Error(t, err3)
		require.Nil(t, d3)
	})
}

func TestDeployer_Advance(t *testing.T) {
	t.Parallel()

	t.Run("advance", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Generate(ctx, "yo", "hello world", "*SYSTEM*")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, Creating, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Gathering, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Merging, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Bundling, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Sending, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Completed, d2.Status())

		d2, err2 = d.Advance()
		require.Error(t, err2)
		require.Nil(t, d2)

		d2, err2 = d.Fail("haha")
		require.Error(t, err2)
		require.Nil(t, d2)
	})
}

func TestDeployer_Advance_Fail(t *testing.T) {
	t.Parallel()

	t.Run("advance fail", func(t *testing.T) {
		t.Parallel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Advance()
		require.Error(t, err2)
		require.Nil(t, d2)
	})
}

func TestDeployer_Fail(t *testing.T) {
	t.Parallel()

	t.Run("fail", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Generate(ctx, "yo", "hello world", "")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, Creating, d2.Status())

		d2, err2 = d.Advance()
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Gathering, d2.Status())

		d2, err2 = d.Fail("not a good run!")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, Failed, d2.Status())

		d2, err2 = d.Advance()
		require.Error(t, err2)
		require.Nil(t, d2)

		d2, err2 = d.Fail("haha")
		require.Error(t, err2)
		require.Nil(t, d2)
	})
}

func TestDeployer_Fail_Fail(t *testing.T) {
	t.Parallel()

	t.Run("fail fail", func(t *testing.T) {
		t.Parallel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Fail("oops")
		require.Error(t, err2)
		require.Nil(t, d2)
	})
}

func TestDeployer_LastDeployment(t *testing.T) {
	t.Parallel()

	t.Run("last deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Generate(ctx, "yo", "hello world", "")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, Creating, d2.Status())

		d3, err3 := d.LastDeployment(ctx)
		require.NoError(t, err3)
		require.NotNil(t, d3)
		assert.Equal(t, d2.Version(), d3.Version())
		assert.Equal(t, d2.Status(), d3.Status())
	})
}

func TestDeployer_LastDeployment_Fail(t *testing.T) {
	t.Parallel()

	t.Run("last deployment fail", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d3, err3 := d.LastDeployment(ctx)
		require.Error(t, err3)
		require.Nil(t, d3)
	})
}

func TestDeployer_ReadDeployment(t *testing.T) {
	t.Parallel()

	t.Run("read deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d2, err2 := d.Generate(ctx, "yo", "hello world", "")
		require.NoError(t, err2)
		require.NotNil(t, d2)
		assert.Equal(t, uint64(1), d2.Version())
		assert.Equal(t, Creating, d2.Status())

		d3, err3 := d.ReadDeployment(ctx, 1)
		require.NoError(t, err3)
		require.NotNil(t, d3)
		assert.Equal(t, d2.Version(), d3.Version())
		assert.Equal(t, d2.Status(), d3.Status())
	})
}

func TestDeployer_ReadDeployment_Fail(t *testing.T) {
	t.Parallel()

	t.Run("read deployment", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		s := memory.New()
		require.NotNil(t, s)

		d := NewKeyValueDB(s, "")
		require.NotNil(t, d)

		d3, err3 := d.ReadDeployment(ctx, 1)
		require.Error(t, err3)
		require.Nil(t, d3)
	})
}

func TestDeployer_ListDeployments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		count int
	}{
		{name: "none", count: 0},
		{name: "one", count: 1},
		{name: "few", count: 5},
		{name: "many", count: 101},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			s := memory.New()
			require.NotNil(t, s)

			d := NewKeyValueDB(s, "")
			require.NotNil(t, d)

			for _ = range tc.count {
				d2, err2 := d.Generate(ctx, "yo", "hello world", "user")
				require.NoError(t, err2)
				require.NotNil(t, d2)

				for {
					d3, err3 := d.Advance()
					require.NoError(t, err3)
					require.NotNil(t, d3)

					if d3.Status() > Sending {
						break
					}
				}
			}

			list, err := d.ListDeployments(ctx)
			require.NoError(t, err)
			require.Equal(t, tc.count, len(list))
		})
	}
}
