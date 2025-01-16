package pip

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/components"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	util "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/slog"
)

func TestParseXML(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "empty", wantErr: true},
		{name: "data1", body: xmlData1},
		{name: "data2", body: xmlData2},
		{name: "data3", body: xmlData3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := models.NewAttributeSet()
			require.NotNil(t, a)

			err := parseXML([]byte(tc.body), a)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

var xmlData1 = `<content>
    <p>this is content area</p>
    <animal>
        <p>This is a dog</p>
        <dog>
           <p>tommy</p>
        </dog>
    </animal>
    <birds>
        <p>this is a bird</p>
        <p>this is a bird</p>
    </birds>
    <animal>
        <p>another animal</p>
    </animal>
</content>`

var xmlData2 = `<data>
    <entry>
        <name>John Doe</name>
        <age>28</age>
    </entry>
    <entry>
        <name>Jane Doe</name>
        <age>29</age>
    </entry>
    <entry>
        <name>Bob Doe</name>
        <age>30</age>
    </entry>
    <entry>
        <name>Beth Doe</name>
        <age>31</age>
    </entry>
</data>`

var xmlData3 = `<page> 
   <title>Apollo 11</title> 
     <redirect title="Foo bar" /> 
     <revision> 
       <text xml:space="preserve"> 
       {{Infobox Space mission 
       |mission_name=&lt;!--See above--&gt; 
       |insignia=Apollo_11_insignia.png}}
       </text> 
     </revision> 
 </page>`

func TestParseJSON(t *testing.T) {
	testCases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{name: "empty"},
		{name: "bad json", body: `this is not [] json`, wantErr: true},
		{name: "array", body: `["hello", "world", 123, true, "well"]`, wantErr: true},
		{name: "object", body: `{"hello": "world", "int": 123, "bool": true}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := models.NewAttributeSet()
			require.NotNil(t, a)

			err := parseJSON([]byte(tc.body), a)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProcessBody(t *testing.T) {
	testCases := []struct {
		name    string
		body    []byte
		attr    models.AttributeSet
		wantLog int
		want    models.AttributeSet
	}{
		{
			name: "empty",
		},
		{
			name:    "no content-type, bad data",
			body:    []byte("hello world"),
			wantLog: 1,
		},
		{
			name:    "no content-type, invalid xml",
			body:    []byte("<xml? wat is dit dan?"),
			wantLog: 1,
		},
		{
			name: "no content-type, xml data",
			body: []byte("<start>hello world</start>"),
			want: models.NewAttributeSet(
				models.NewAttribute(
					"body", []map[string]any{
						{"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "hello world",
							"nodes":      []map[string]any{},
						}}},
				),
			),
		},
		{
			name:    "no content-type, invalid json",
			body:    []byte(`{"veld1": wat is dit dan?`),
			wantLog: 1,
		},
		{
			name: "no content-type, json data",
			body: []byte(`{"hello": "world"}`),
			want: models.NewAttributeSet(models.NewAttribute("body", map[string]any{"hello": "world"})),
		},
		{
			name:    "content-type unsupported",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "x-iets")),
			body:    []byte("[]"),
			wantLog: 1,
		},
		{
			name:    "content-type xml, invalid xml",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "application/xml")),
			body:    []byte("<xml? wat is dit dan?"),
			wantLog: 1,
		},
		{
			name: "content-type xml, valid xml",
			attr: models.NewAttributeSet(models.NewAttribute("content-type", "application/xml")),
			body: []byte("<start>hello world</start>"),
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "application/xml"),
				models.NewAttribute(
					"body", []map[string]any{
						{"start": map[string]any{
							"attributes": []map[string]any{},
							"cdata":      "hello world",
							"nodes":      []map[string]any{},
						}}},
				),
			),
		},
		{
			name:    "content-type json, invalid json",
			attr:    models.NewAttributeSet(models.NewAttribute("content-type", "application/json")),
			body:    []byte("[nee,toch}"),
			wantLog: 1,
		},
		{
			name: "content-type json, valid json",
			attr: models.NewAttributeSet(models.NewAttribute("content-type", "text/json")),
			body: []byte(`{"hello": "world"}`),
			want: models.NewAttributeSet(
				models.NewAttribute("content-type", "text/json"),
				models.NewAttribute("body", map[string]any{"hello": "world"}),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := models.NewAttributeSet(tc.attr)
			require.NotNil(t, a)

			h := util.NewDummyHandler(slog.LevelDebug)

			p := New(Config{Logger: slog.New(h)})
			require.NotNil(t, p)

			h.Clear()

			req := &components.Request{Body: tc.body}

			p2, ok := p.(*pip)
			require.True(t, ok)
			require.NotNil(t, p2)
			p2.decodeBody(req, a)

			assert.Equal(t, tc.wantLog, h.Count())

			if tc.want != nil {
				tc.want.IterateAttributes(func(attr models.Attribute) {
					v2 := a.GetAttributeValue(attr.Key())
					assert.EqualValues(t, attr.Value(), v2)
				})

				a.IterateAttributes(func(attr models.Attribute) {
					v2 := tc.want.GetAttributeValue(attr.Key())
					assert.EqualValues(t, attr.Value(), v2)
				})
			}
		})
	}
}
