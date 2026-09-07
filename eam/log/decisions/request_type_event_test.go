package decisions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthRequestTypeEventName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "adl.access_evaluation", EvaluationEndpoint.EventName())
	assert.Equal(t, "adl.access_evaluations", EvaluationsEndpoint.EventName())
	assert.Equal(t, "adl.search_subject", SearchSubjectEndpoint.EventName())
	assert.Equal(t, "adl.search_action", SearchActionEndpoint.EventName())
	assert.Equal(t, "adl.search_resource", SearchResourceEndpoint.EventName())
}
