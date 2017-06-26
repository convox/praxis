package manifest_test

import (
	"testing"

	"github.com/convox/praxis/manifest"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowSteps(t *testing.T) {
	m, err := testdataManifest("full", manifest.Environment{"SECRET": "shh"})
	if !assert.NoError(t, err) {
		return
	}

	wf := m.Workflows.Find("change", "close")

	assert.Equal(t, "change", wf.Type)
	assert.Equal(t, "close", wf.Trigger)

	if assert.Len(t, wf.Steps, 1) {
		assert.Equal(t, "delete", wf.Steps[0].Type)
		assert.Equal(t, "staging/praxis-$branch", wf.Steps[0].Target)
	}

	assert.Nil(t, m.Workflows.Find("foo", "bar"))
}
