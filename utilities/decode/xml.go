package decode

import "encoding/xml"

type xmlNode struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Data    string     `xml:",innerxml"`
	Nodes   []xmlNode  `xml:",any"`
}

func (n *xmlNode) name() string {
	return n.XMLName.Local
}

func (n *xmlNode) toMap() map[string]any {
	attrs := make([]map[string]any, len(n.Attrs))
	for i := range n.Attrs {
		a := n.Attrs[i]
		attrs[i] = map[string]any{a.Name.Local: a.Value}
	}

	nodes := make([]map[string]any, len(n.Nodes))
	for i := range n.Nodes {
		node := n.Nodes[i]
		nodes[i] = map[string]any{node.name(): node.toMap()}
	}

	var cdata string
	if len(n.Nodes) == 0 && len(n.Data) > 0 {
		cdata = n.Data
	}

	return map[string]any{
		"attributes": attrs,
		"nodes":      nodes,
		"cdata":      cdata,
	}
}

// UnmarshalXML implements the xml.Unmarshaler interface.
func (n *xmlNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	n.Attrs = start.Attr
	type node xmlNode
	return d.DecodeElement((*node)(n), &start)
}
