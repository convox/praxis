package local_test

import (
	"testing"

	"github.com/convox/praxis/types"
	"github.com/stretchr/testify/assert"
)

func TestReleaseCreateGet(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	_, err = p.AppCreate("app")
	assert.NoError(t, err)

	opts := types.ReleaseCreateOptions{
		Build: "BTEST",
		Env: map[string]string{
			"APP": "app",
			"FOO": "bar",
		},
	}
	rel, err := p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	fetched, err := p.ReleaseGet("app", rel.Id)
	assert.NoError(t, err)

	assert.Equal(t, rel, fetched)
}

func TestReleaseList(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	_, err = p.AppCreate("app")
	assert.NoError(t, err)

	opts := types.ReleaseCreateOptions{
		Build: "BTEST",
		Env: map[string]string{
			"APP": "app",
			"FOO": "bar",
		},
	}
	rel1, err := p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	opts.Build = "BTEST2"
	rel2, err := p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	opts.Build = "BTEST3"
	opts.Env["FOO"] = "baz"
	rel3, err := p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	rels, err := p.ReleaseList("app")
	assert.Len(t, rels, 3)

	rel1.Env = map[string]string{"APP": "app", "FOO": "bar"}
	rel2.Env = map[string]string{"APP": "app", "FOO": "bar"}
	rel3.Env = map[string]string{"APP": "app", "FOO": "baz"}

	assert.Equal(t, *rel3, rels[0])
	assert.Equal(t, *rel2, rels[1])
	assert.Equal(t, *rel1, rels[2])
}
