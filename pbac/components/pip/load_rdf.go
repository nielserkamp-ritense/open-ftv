package pip

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/deiu/rdf2go"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/pbac/models"
	mime "gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/io"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/rdf"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/ftv-implementatie/utilities/turtle"
)

func (p *pip) loadRDF(f io.Reader, path, mt string) {
	if f == nil {
		p.logger.Error("pip: nil input", "path", path, "mimetype", mt)
		return
	}

	if mt == "" {
		mt = mime.MimeTypeTurtle
	}

	graph, err := turtle.Load(f, mt)
	if err != nil {
		p.logger.Error("pip: error decoding RDF file", "path", path, "mimetype", mt, "err", err)
		return
	}

	l := &loader{path: path, mt: mt, logger: p.logger, graph: graph, p: p}
	l.run()
}

func (l *loader) run() {
	l.loadEntities()
	l.loadAttributes()
}

func (l *loader) loadEntities() {
	list := l.graph.All(nil, rdf2go.NewResource(rdf.Type), rdf2go.NewResource(rdf.FTVEntity))
	for i := range list {
		sub := list[i].Subject

		ns, id, err := l.getTypeID(sub)
		if err != nil {
			l.logger.Error("pip: error processing RDF entities", "path", l.path, "mimetype", l.mt, "err", err)
		}

		attributes := l.p.newAttributes()

		attrList := l.graph.All(sub, rdf2go.NewResource(rdf.FTVEntityAttribute), nil)
		for j := range attrList {
			k, v, err2 := l.getKeyValue(attrList[j].Object)
			if err2 != nil {
				l.logger.Error("pip: error processing RDF entity attributes", "path", l.path, "mimetype", l.mt, "err", err2)
			} else {
				attributes.AddAttribute(k, v)
			}
		}

		l.p.AddEntity(models.NewEntity(ns, id, attributes))
	}
}

func (l *loader) loadAttributes() {
	list := l.graph.All(nil, rdf2go.NewResource(rdf.Type), rdf2go.NewResource(rdf.FTVAttribute))
	for i := range list {
		k, v, err := l.getKeyValue(list[i].Subject)
		if err != nil {
			l.logger.Error("pip: error processing RDF attributes", "path", l.path, "mimetype", l.mt, "err", err)
		} else {
			l.p.AddAttribute(k, v)
		}
	}
}

func (l *loader) getTypeID(sub rdf2go.Term) (string, string, error) {
	t, err1 := l.getString(sub, rdf.FTVEntityType)
	if err1 != nil {
		return "", "", err1
	}

	id, err2 := l.getString(sub, rdf.FTVEntityID)
	if err2 != nil {
		return "", "", err2
	}

	return t, id, nil
}

func (l *loader) getKeyValue(sub rdf2go.Term) (string, any, error) {
	k, err1 := l.getString(sub, rdf.FTVAttributeKey)
	if err1 != nil {
		return "", nil, err1
	}

	v, err2 := l.getValue(sub, rdf.FTVAttributeValue)
	if err2 != nil {
		return "", nil, err2
	}

	return k, v, nil
}

func (l *loader) getString(sub rdf2go.Term, pred string) (string, error) {
	a, err1 := l.getValue(sub, pred)
	if err1 != nil {
		return "", err1
	}

	k, ok := a.(string)
	if !ok {
		return "", fmt.Errorf("pip: invalid key type (%v)", a)
	}

	return k, nil
}

func (l *loader) getValue(sub rdf2go.Term, pred string) (any, error) {
	list := l.graph.All(sub, rdf2go.NewResource(pred), nil)
	if len(list) == 0 {
		return nil, fmt.Errorf("pip: no objects found")
	}
	return l.convertValues(list)
}

func (l *loader) convertValues(list []*rdf2go.Triple) (any, error) {
	if l.allLiteral(list) {
		return l.convertLiterals(list)
	}
	return l.convertObjects(list)
}

func (l *loader) allLiteral(list []*rdf2go.Triple) bool {
	for i := range list {
		if _, ok := list[i].Object.(*rdf2go.Literal); !ok {
			return false
		}
	}
	return true
}

func (l *loader) convertLiterals(list []*rdf2go.Triple) (any, error) {
	out := make([]any, 0, len(list))

	for i := range list {
		a, err := l.convertLiteral(list[i].Object)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}

	if len(out) == 1 {
		return out[0], nil
	}
	return out, nil
}

func (l *loader) convertLiteral(obj rdf2go.Term) (any, error) {
	switch t := obj.(type) {
	case *rdf2go.Literal:
		return l.getLiteral(t)
	default:
		return nil, fmt.Errorf("pip: invalid object type (%v)", t)
	}
}

func (l *loader) convertObjects(list []*rdf2go.Triple) (any, error) {
	out := make(map[string]any, len(list))

	var id int

	for i := range list {
		k, v, err := l.convertObject(list[i].Object)
		if err != nil {
			return nil, err
		}

		if k == "" {
			id++
			k = fmt.Sprintf("__obj_%d", id)
		}
		out[k] = v
	}

	return out, nil
}

func (l *loader) convertObject(obj rdf2go.Term) (string, any, error) {
	switch t := obj.(type) {
	case *rdf2go.Literal:
		v, err := l.getLiteral(t)
		return "", v, err

	case *rdf2go.BlankNode, *rdf2go.Resource:
		k, err1 := l.getString(obj, rdf.FTVAttributeKey)
		if err1 != nil {
			return "", nil, err1
		}

		v, err := l.getValue(obj, rdf.FTVAttributeValue)
		if err != nil {
			v, err = l.getValue(obj, rdf.FTVEntityAttribute)
			if err != nil {
				return k, nil, nil
			}
		}
		return k, v, nil

	default:
		return "", nil, fmt.Errorf("pip: invalid object type (%v)", t)
	}
}

func (l *loader) getLiteral(obj *rdf2go.Literal) (any, error) {
	if d := obj.Datatype; d != nil {
		return rdf.ConvertLiteral(obj.Value, d.RawValue())
	}
	return rdf.ConvertLiteral(obj.Value, rdf.XSDString)
}

type loader struct {
	path   string
	mt     string
	logger *slog.Logger
	graph  *rdf2go.Graph
	p      *pip
}
