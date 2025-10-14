package postgresql

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/decisions"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/eam/log/search"
	oas "gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authlog"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/authzen"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/convert"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql"
	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/utilities/storage/postgresql/pool"
)

// Decisions implements an ADL search tool with a PostgrSQL database as backend.
type Decisions struct {
	db *postgresql.Postgres
}

// New instantiates a new ADL search tool with a PostgrSQL database as backend.
func New(ctx context.Context, dsn string, p *pool.Pool, maxLife time.Duration, maxConn int32) (*Decisions, error) {
	var db *postgresql.Postgres
	var err error

	if p == nil {
		db, err = postgresql.New(ctx, dsn, maxLife, maxConn)
	} else {
		db, err = postgresql.NewWithPool(ctx, dsn, maxLife, maxConn, p)
	}

	if err != nil {
		return nil, err
	}
	return &Decisions{db: db}, nil
}

// Search performs a query on the PostgreSQL database and returns records matching the given search criteria.
func (d *Decisions) Search(ctx context.Context, criteria *search.Criteria) (out oas.AuthlogEntries, err error) {
	if err = criteria.Test(); err != nil {
		return
	}

	needMatching := criteria.SubjectType != "" || criteria.SubjectId != "" ||
		criteria.ActionName != "" || criteria.ResourceType != "" || criteria.ResourceId != ""

	sql, params := buildQuery(criteria, needMatching)
	_, _ = sql, params

	out = make(oas.AuthlogEntries, 0)

	if err = d.db.Query(ctx, sql, params, func(values []any) bool {
		if !needMatching || criteriaMatch(criteria, values) {
			out = append(out, buildLogRecord(values))
		}
		return len(out) < criteria.Limit
	}); err != nil {
		return nil, err
	}

	return
}

func buildQuery(criteria *search.Criteria, needMatching bool) (string, []any) {
	var sql bytes.Buffer
	var params []any

	sql.WriteString("SELECT ")
	sql.WriteString("id,created,request_type,policies,request,response,information,engine,trace_id,span_id")
	sql.WriteString(" FROM decision")

	sql.WriteString(" WHERE")

	if criteria.ID > 0 {
		sql.WriteString(" id=$1")
		params = append(params, criteria.ID)
	} else {
		sql.WriteString(" created>=$1 AND created<=$2")
		params = append(params, criteria.From, criteria.To)

		switch len(criteria.RequestTypes) {
		case 0:
			// no-op
		case 1:
			sql.WriteString(fmt.Sprintf(" AND request_type=$%d", len(params)+1))
			params = append(params, criteria.RequestTypes[0])
		default: // more than 1
			sql.WriteString(" AND request_type IN [")
			for i := range criteria.RequestTypes {
				if i > 0 {
					sql.WriteByte(',')
				}
				sql.WriteString(fmt.Sprintf("$%d", len(params)+1))
				params = append(params, criteria.RequestTypes[i])
			}
			sql.WriteByte(']')
		}

		switch len(criteria.Bundles) {
		case 0:
			// no-op
		case 1:
			sql.WriteString(fmt.Sprintf(" AND policies=$%d", len(params)+1))
			params = append(params, criteria.Bundles[0])
		default: // more than 1
			sql.WriteString(" AND policies IN [")
			for i := range criteria.Bundles {
				if i > 0 {
					sql.WriteByte(',')
				}
				sql.WriteString(fmt.Sprintf("$%d", len(params)+1))
				params = append(params, criteria.Bundles[i])
			}
			sql.WriteByte(']')
		}

		if b := criteria.GetTraceID(); b != nil {
			sql.WriteString(fmt.Sprintf(" AND trace_id=$%d", len(params)+1))
			params = append(params, b)
		}

		if b := criteria.GetSpanID(); b != nil {
			sql.WriteString(fmt.Sprintf(" AND span_id=$%d", len(params)+1))
			params = append(params, b)
		}

		sql.WriteString(" ORDER BY created DESC")

		if !needMatching {
			sql.WriteString(fmt.Sprintf(" LIMIT $%d", len(params)+1))
			params = append(params, criteria.Limit)
		}
	}

	return sql.String(), params
}

func criteriaMatch(criteria *search.Criteria, values []any) bool {
	rt := decisions.AuthRequestType(convert.AnyToUint64(values[2]))

	switch rt {
	case decisions.EvaluationEndpoint:
		b := anyToJSON(values[4])
		req := new(authzen.EvaluationRequest)
		if err := json.Unmarshal(b, req); err == nil {
			return evaluationMatch(criteria, req)
		}

	case decisions.EvaluationsEndpoint:
		b := anyToJSON(values[4])
		req := new(authzen.EvaluationsRequest)
		if err := json.Unmarshal(b, req); err == nil {
			return evaluationsMatch(criteria, req)
		}

	case decisions.SearchSubjectEndpoint, decisions.SearchResourceEndpoint:
		b := anyToJSON(values[4])
		req := new(authzen.SearchRequest)
		if err := json.Unmarshal(b, req); err == nil {
			return searchMatch(criteria, req)
		}

	case decisions.SearchActionEndpoint:
		b := anyToJSON(values[4])
		req := new(authzen.SearchActionRequest)
		if err := json.Unmarshal(b, req); err == nil {
			return searchActionMatch(criteria, req)
		}
	}

	return false
}

func evaluationMatch(criteria *search.Criteria, request *authzen.EvaluationRequest) bool {
	return sarMatch(criteria, &request.Subject, &request.Resource, &request.Action)
}

func evaluationsMatch(criteria *search.Criteria, request *authzen.EvaluationsRequest) bool {
	if sarMatch(criteria, &request.Subject, &request.Resource, &request.Action) {
		return true
	}

	for i := range request.Evaluations {
		e := &request.Evaluations[i]
		if sarMatch(criteria, &e.Subject, &e.Resource, &e.Action) {
			return true
		}
	}

	return false
}

func searchMatch(criteria *search.Criteria, request *authzen.SearchRequest) bool {
	return sarMatch(criteria, authzen.EntityFromSearch(&request.Subject), authzen.EntityFromSearch(&request.Resource), &request.Action)
}

func searchActionMatch(criteria *search.Criteria, request *authzen.SearchActionRequest) bool {
	return sarMatch(criteria, &request.Subject, &request.Resource, &authzen.Action{})
}

func sarMatch(criteria *search.Criteria, subject, resource *authzen.Entity, action *authzen.Action) bool {
	return (criteria.SubjectType == "" || subject.Type == criteria.SubjectType) &&
		(criteria.SubjectId == "" || subject.Id == criteria.SubjectId) &&
		(criteria.ActionName == "" || action.Name == criteria.ActionName) &&
		(criteria.ResourceType == "" || resource.Type == criteria.ResourceType) &&
		(criteria.ResourceId == "" || resource.Id == criteria.ResourceId)
}

func buildLogRecord(values []any) oas.AuthlogEntry {
	return oas.AuthlogEntry{
		Id:          convert.AnyToInt64(values[0]),
		Created:     convert.AnyToDateTime(values[1]).Format(time.RFC3339),
		RequestType: decisions.AuthRequestType(convert.AnyToInt64(values[2])).String(),
		Policies:    convert.AnyToString(values[3]),
		Request:     anyToMap(values[4]),
		Response:    anyToMap(values[5]),
		Information: anyToMap(values[6]),
		Engine:      anyToMap(values[7]),
		TraceId:     anyToHex(values[8]),
		SpanId:      anyToHex(values[9]),
	}
}

func anyToMap(in any) map[string]any {
	if m, ok := in.(map[string]any); ok {
		return m
	}
	return nil
}

func anyToHex(in any) string {
	if b, ok := in.([]byte); ok {
		return hex.EncodeToString(b)
	}
	return ""
}

func anyToJSON(in any) []byte {
	if m := anyToMap(in); m != nil {
		if b, err := json.Marshal(m); err == nil {
			return b
		}
	}
	return []byte("{}")
}
