package bundles

import (
	"testing"
	"time"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeployment(t *testing.T) {
	t.Parallel()

	started := time.Now().UTC()

	testCases := []struct {
		name    string
		version uint64
		title   string
		desc    string
	}{
		{name: "v1", version: 1, title: "v1", desc: "version 1"},
		{name: "v101", version: 101, title: "v101", desc: "blah blah"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := NewDeployment(tc.version, tc.title, tc.desc, "")
			require.NotNil(t, d)
			assert.Equal(t, tc.version, d.Version())
			assert.Equal(t, Creating, d.Status())
			assert.LessOrEqual(t, d.created, time.Now().UTC())
			assert.LessOrEqual(t, d.updated, time.Now().UTC())
			assert.GreaterOrEqual(t, d.created, started)
			assert.GreaterOrEqual(t, d.updated, started)

			for i := Gathering; i < Failed; i++ {
				s := d.NextStatus()
				assert.Equal(t, i, s)
			}

			s := d.Failed("oopsie")
			assert.Equal(t, Failed, s)
			assert.Equal(t, "oopsie", d.msg)

			s2 := d.Completed()
			assert.Equal(t, Failed, s2)
			assert.Equal(t, "oopsie", d.msg)

			d2 := NewDeployment(tc.version, tc.title, tc.desc, "")
			require.NotNil(t, d2)

			s3 := d2.Completed()
			assert.Equal(t, Completed, s3)
			assert.Empty(t, d2.msg)

			s4 := d2.Failed("oopsie")
			assert.Equal(t, Completed, s4)
			assert.Empty(t, d2.msg)
		})
	}
}

func TestDeployment_Failed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		version uint64
		title   string
		desc    string
		status  Status
		msg     string
		wantErr bool
	}{
		{name: "v1", version: 1, title: "v1", desc: "version 1", status: Creating, msg: "oopsie1"},
		{name: "v101", version: 101, title: "v101", desc: "blah blah", status: Gathering, msg: "oopsie2"},
		{name: "v102", version: 102, title: "v102", desc: "blah blah", status: Failed, msg: "oopsie3", wantErr: true},
		{name: "v103", version: 103, title: "v103", desc: "blah blah", status: Completed, msg: "oopsie4", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := NewDeployment(tc.version, tc.title, tc.desc, "")
			require.NotNil(t, d)
			assert.Equal(t, tc.version, d.Version())

			d.status = tc.status

			s := d.Failed(tc.msg)
			if tc.wantErr {
				assert.Equal(t, tc.status, s)
				assert.Empty(t, d.msg)
			} else {
				assert.Equal(t, Failed, s)
				assert.Equal(t, tc.msg, d.msg)
			}
		})
	}
}

func TestDeployment_Completed(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		version uint64
		title   string
		desc    string
		status  Status
		wantErr bool
	}{
		{name: "v1", version: 1, title: "v1", desc: "version 1", status: Creating},
		{name: "v101", version: 101, title: "v101", desc: "blah blah", status: Gathering},
		{name: "v102", version: 102, title: "v102", desc: "blah blah", status: Failed, wantErr: true},
		{name: "v103", version: 103, title: "v103", desc: "blah blah", status: Completed, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := NewDeployment(tc.version, tc.title, tc.desc, "")
			require.NotNil(t, d)
			assert.Equal(t, tc.version, d.Version())

			d.status = tc.status

			s := d.Completed()
			if tc.wantErr {
				assert.Equal(t, tc.status, s)
			} else {
				assert.Equal(t, Completed, s)
			}

			assert.Empty(t, d.msg)
		})
	}
}

func TestDeployment_JSON(t *testing.T) {
	t.Parallel()

	t1 := time.Date(2025, 7, 22, 12, 13, 14, 0, time.UTC)

	testCases := []struct {
		name    string
		version uint64
		title   string
		desc    string
		want    string
	}{
		{
			name:    "v1",
			version: 1,
			title:   "v1",
			desc:    "version 1",
			want:    `{"audit":{"created":"2025-07-22T12:13:14Z","createdBy":"","updated":"2025-07-22T12:13:14Z"},"description":"version 1","status":1,"title":"v1","version":1}`,
		},
		{
			name:    "v101",
			version: 101,
			title:   "v101",
			desc:    "blah blah",
			want:    `{"audit":{"created":"2025-07-22T12:13:14Z","createdBy":"","updated":"2025-07-22T12:13:14Z"},"description":"blah blah","status":1,"title":"v101","version":101}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d := NewDeployment(tc.version, tc.title, tc.desc, "")
			require.NotNil(t, d)

			d.created = t1
			d.updated = t1

			got, err := json.Marshal(d)
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, tc.want, string(got))

			var d2 Deployment
			err = json.Unmarshal(got, &d2)
			require.NoError(t, err)
			assert.Equal(t, d.Version(), d2.Version())
			assert.Equal(t, d.Status(), d2.Status())
			assert.Equal(t, d.description, d2.description)
			assert.Equal(t, d.msg, d2.msg)
			assert.Equal(t, d.created, d2.created)
			assert.Equal(t, d.updated, d2.updated)
		})
	}
}
