package adl

import (
	"context"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/models"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
)

// New instantiates an FTV Authorization Decision Log writer.
func New(logger *decisions.Logger, opts ...Option) *ADL {
	l := &ADL{logger: logger}
	for i := range opts {
		opts[i](l)
	}
	return l
}

// Evaluation logs a new evaluation call.
func (l *ADL) Evaluation(ctx context.Context, timestamp time.Time, req *oas.EvaluationRequest, resp *oas.EvaluationResponse, status decisions.Status) error {
	return l.write(ctx, timestamp, decisions.EvaluationEndpoint, req, resp, status)
}

// Evaluations logs a new evaluations call.
func (l *ADL) Evaluations(ctx context.Context, timestamp time.Time, req *oas.EvaluationsRequest, resp *oas.EvaluationsResponse, status decisions.Status) error {
	return l.write(ctx, timestamp, decisions.EvaluationsEndpoint, req, resp, status)
}

// SearchSubject logs a new subject search call.
func (l *ADL) SearchSubject(ctx context.Context, timestamp time.Time, req *oas.SearchRequest, resp *oas.SearchResponse, status decisions.Status) error {
	return l.write(ctx, timestamp, decisions.SearchSubjectEndpoint, req, resp, status)
}

// SearchAction logs a new action search call.
func (l *ADL) SearchAction(ctx context.Context, timestamp time.Time, req *oas.SearchActionRequest, resp *oas.SearchActionResponse, status decisions.Status) error {
	return l.write(ctx, timestamp, decisions.SearchActionEndpoint, req, resp, status)
}

// SearchResource logs a new resource search call.
func (l *ADL) SearchResource(ctx context.Context, timestamp time.Time, req *oas.SearchRequest, resp *oas.SearchResponse, status decisions.Status) error {
	return l.write(ctx, timestamp, decisions.SearchResourceEndpoint, req, resp, status)
}

func (l *ADL) write(ctx context.Context, timestamp time.Time, tr decisions.AuthRequestType, req, resp any, status decisions.Status) error {
	traceID := convert.AnyToString(ctx.Value(models.AttrTraceID))
	spanID := convert.AnyToString(ctx.Value(models.AttrSpanID))
	parentSpanID := convert.AnyToString(ctx.Value(models.AttrParentSpanID))

	return l.logger.Decision(ctx, &decisions.Decision{
		Timestamp:    timestamp.UTC(),
		RequestType:  tr,
		EventName:    tr.EventName(),
		Status:       status,
		Request:      req,
		Response:     resp,
		Policies:     l.bundleVersion,
		Information:  l.information,
		Engine:       l.engine,
		Resource:     l.resource,
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
	})
}

// Shutdown flushes any queued decisions and releases resources held by the
// underlying decision log.
//
// Do not call Evaluation/Evaluations/Search* after Shutdown has been called.
func (l *ADL) Shutdown(ctx context.Context) error {
	if l == nil || l.logger == nil {
		return nil
	}

	return l.logger.Shutdown(ctx)
}

// NewBundle associates a new bundle version with the ADL.
func (l *ADL) NewBundle(version uint64) {
	l.bundleVersion = version
}

// NewInformation associates new information with the ADL.
func (l *ADL) NewInformation(information map[string]any) {
	if len(information) == 0 {
		l.information = nil
		return
	}

	b, _ := json.Marshal(information)
	l.information = b
}

// NewEngine associates new policy-engine details with the ADL.
func (l *ADL) NewEngine(engine map[string]any) {
	if len(engine) == 0 {
		l.engine = nil
		return
	}

	b, _ := json.Marshal(engine)
	l.engine = b
}

// ADL implements an FTV Authorization Decision Log writer.
type ADL struct {
	logger        *decisions.Logger
	bundleVersion uint64
	information   any
	engine        any
	resource      any
}
