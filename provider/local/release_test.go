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

	if assert.NotNil(t, rel) {
		fetched, err := p.ReleaseGet("app", rel.Id)
		assert.NoError(t, err)

		assert.Equal(t, rel, fetched)
	}
}

func TestReleaseList(t *testing.T) {
	p, err := testProvider()
	assert.NoError(t, err)
	defer cleanup(p)

	_, err = p.AppCreate("app")
	assert.NoError(t, err)

	expects := types.Releases{
		{App: "app", Build: "BTEST3", Env: map[string]string{"APP": "app", "FOO": "baz"}},
		{App: "app", Build: "BTEST2", Env: map[string]string{"APP": "app", "FOO": "bar"}},
		{App: "app", Build: "BTEST", Env: map[string]string{"APP": "app", "FOO": "bar"}},
	}

	opts := types.ReleaseCreateOptions{
		Build: "BTEST",
		Env: map[string]string{
			"APP": "app",
			"FOO": "bar",
		},
	}
	_, err = p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	opts.Build = "BTEST2"
	_, err = p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	opts.Build = "BTEST3"
	opts.Env["FOO"] = "baz"
	_, err = p.ReleaseCreate("app", opts)
	assert.NoError(t, err)

	rels, err := p.ReleaseList("app")
	assert.Len(t, rels, 3)

	for i := range rels {
		assert.Equal(t, expects[i].App, rels[i].App)
		assert.Equal(t, expects[i].Build, rels[i].Build)
		assert.Equal(t, expects[i].Env, rels[i].Env)
		assert.Equal(t, false, rels[i].Created.IsZero())
		assert.NotEqual(t, "", rels[i].Id)
	}
}
