package fiber

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/gofiber/fiber/v2"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/principals"
)

// principalTypeName is the name every OAS spec generates for its attribution object. Each spec
// produces its own Go type, so the walk below matches on shape rather than on a single type.
const principalTypeName = "Principal"

// maxPrincipalDepth bounds the response walk. Responses are shallow; the limit only stops a
// pathological or cyclic value from spinning.
const maxPrincipalDepth = 12

// HandlerOption configures a handler with optional collaborators.
type HandlerOption func(*principalResolver)

// WithPrincipals gives a handler the principal store it uses to turn the attribution ids in a
// response into names. Without it a response carries ids only, and clients render those.
func WithPrincipals(store principals.Store) HandlerOption {
	return func(r *principalResolver) {
		r.store = store
	}
}

// principalResolver turns the stored attribution ids in a response into display data.
//
// Resolution happens here, at the serialisation boundary, rather than as a SQL join in each read
// query: the principal table belongs to eam/principals, and the components that own policies,
// entities, settings and bundles should not reach into it. See docs/adr/0004.
type principalResolver struct {
	store  principals.Store
	logger *slog.Logger
}

func newPrincipalResolver(logger *slog.Logger, opts []HandlerOption) principalResolver {
	r := principalResolver{logger: logger}

	for i := range opts {
		opts[i](&r)
	}

	return r
}

// respond resolves every attribution reference in v and writes it as JSON.
//
// It walks the response rather than being called per field, so an endpoint added later cannot
// forget to resolve and silently serve raw ids.
//
// A resolution failure is logged and the response sent unresolved: names are display data, and a
// read must not fail because the principal store is unavailable.
func (r principalResolver) respond(req *fiber.Ctx, v any) error {
	if r.store == nil || v == nil {
		return req.JSON(v)
	}

	// Box v so the walk always has an addressable value: a response passed by value would
	// otherwise be unsettable and silently resolve to nothing.
	box := reflect.New(reflect.TypeOf(v))
	box.Elem().Set(reflect.ValueOf(v))

	if err := r.resolve(req.Context(), box); err != nil && r.logger != nil {
		r.logger.Warn("failed to resolve principals; responding with ids only", "path", req.Path(), "err", err)
	}

	return req.JSON(box.Elem().Interface())
}

// resolve fills in display data for every Principal reachable from v, in a single batched lookup.
//
// A reference whose id has no record is left untouched, so the response keeps the raw id and the
// client shows that instead. That is the fallback that lets attribution survive a principal the
// manager has never seen -- including every *_audit row, which carries no foreign key.
func (r principalResolver) resolve(ctx context.Context, v reflect.Value) error {
	var found []reflect.Value

	collectPrincipals(v, 0, &found)

	if len(found) == 0 {
		return nil
	}

	ids := make([]string, 0, len(found))
	for _, p := range found {
		ids = append(ids, p.FieldByName("Id").String())
	}

	records, err := r.store.Resolve(ctx, ids)
	if err != nil {
		return err
	}

	for _, p := range found {
		rec, ok := records[p.FieldByName("Id").String()]
		if !ok {
			continue
		}

		setStringField(p, "Name", rec.Name)
		setStringField(p, "Kind", rec.Kind)
	}

	return nil
}

// collectPrincipals appends every settable Principal struct reachable from v to out.
func collectPrincipals(v reflect.Value, depth int, out *[]reflect.Value) {
	if depth > maxPrincipalDepth || !v.IsValid() {
		return
	}

	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if !v.IsNil() {
			collectPrincipals(v.Elem(), depth+1, out)
		}

	case reflect.Slice, reflect.Array:
		// Skip element types that cannot contain a Principal. Without this, a []byte logo or an
		// arbitrary attribute payload is walked one element at a time on every response.
		if !mayContainPrincipal(v.Type().Elem()) {
			return
		}

		for i := range v.Len() {
			collectPrincipals(v.Index(i), depth+1, out)
		}

	case reflect.Map:
		for _, k := range v.MapKeys() {
			collectPrincipals(v.MapIndex(k), depth+1, out)
		}

	case reflect.Struct:
		if isPrincipal(v) {
			*out = append(*out, v)
			return
		}

		for i := range v.NumField() {
			if v.Type().Field(i).IsExported() {
				collectPrincipals(v.Field(i), depth+1, out)
			}
		}

	default:
	}
}

// mayContainPrincipal reports whether a value of type t could transitively hold a Principal.
// Only composite types can; everything else (notably []byte and string) is a dead end.
func mayContainPrincipal(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Struct, reflect.Map, reflect.Interface:
		return true
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return mayContainPrincipal(t.Elem())
	default:
		return false
	}
}

// isPrincipal reports whether v is a generated OAS Principal that this walk may write to.
func isPrincipal(v reflect.Value) bool {
	if v.Type().Name() != principalTypeName || !v.CanSet() {
		return false
	}

	id := v.FieldByName("Id")

	return id.IsValid() && id.Kind() == reflect.String
}

// setStringField assigns to a string-kinded field, tolerating the named string types the codegen
// produces for enums (PrincipalKind).
func setStringField(v reflect.Value, name, value string) {
	f := v.FieldByName(name)
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
		f.SetString(value)
	}
}
