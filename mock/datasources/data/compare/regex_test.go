package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRX(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      string
		wantErr bool
		wantRX  string
	}{
		{name: "bad", in: "[)[", wantErr: true},
		{name: "plain", in: "ha.*", wantRX: "^ha.*$"},
		{name: "with prefix", in: "^hello", wantRX: "^hello$"},
		{name: "with suffix", in: "world$", wantRX: "^world$"},
		{name: "with both", in: "^ha.*$", wantRX: "^ha.*$"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := buildRX(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantRX, got.String())
			}
		})
	}
}

func TestRXFromString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      any
		wantErr bool
		wantRX  string
	}{
		{name: "bad", in: "[)[", wantErr: true},
		{name: "integer", in: 123, wantRX: "^123$"},
		{name: "bool", in: true, wantRX: "^true$"},
		{name: "float", in: 123.45, wantRX: "^123.45$"},
		{name: "plain", in: "ha.*", wantRX: "^ha.*$"},
		{name: "with prefix", in: "^hello", wantRX: "^hello$"},
		{name: "with suffix", in: "world$", wantRX: "^world$"},
		{name: "with both", in: "^ha.*$", wantRX: "^ha.*$"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := RXFromString(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantRX, got.String())
			}
		})
	}
}

func TestRXFromLike(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		in      any
		wantErr bool
		wantRX  string
	}{
		{name: "integer", in: 123, wantRX: "^123$"},
		{name: "bool", in: true, wantRX: "^true$"},
		{name: "float", in: 123.45, wantRX: "^123\\.45$"},
		{name: "dots", in: "...", wantRX: "^\\.\\.\\.$"},
		{name: "specials", in: "+()[]{}", wantRX: "^\\+\\(\\)\\[\\]\\{\\}$"},
		{name: "like", in: "%lo%", wantRX: "^.*lo.*$"},
		{name: "plain", in: "ha.*", wantRX: "^ha\\..*$"},
		{name: "with prefix", in: "^hello", wantRX: "^hello$"},
		{name: "with suffix", in: "world?$", wantRX: "^world.$"},
		{name: "with both", in: "^ha.*$", wantRX: "^ha\\..*$"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := RXFromLike(tc.in)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tc.wantRX, got.String())
			}
		})
	}
}
