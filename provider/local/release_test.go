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
	if !assert.NoError(t, err) {
		return
	}

	p.ReleaseCreate("app", types.ReleaseCreateOptions{Build: "B1"})
	p.ReleaseCreate("app", types.ReleaseCreateOptions{Env: map[string]string{"FOO": "bar"}})
	p.ReleaseCreate("app", types.ReleaseCreateOptions{Build: "B2"})
	p.ReleaseCreate("app", types.ReleaseCreateOptions{Build: "B3"})
	p.ReleaseCreate("app", types.ReleaseCreateOptions{Env: map[string]string{"FOO": "baz"}})
	p.ReleaseCreate("app", types.ReleaseCreateOptions{Build: "B4"})

	rs, err := p.ReleaseList("app")

	if assert.NoError(t, err) && assert.Len(t, rs, 6) {
		assert.Equal(t, "B4", rs[0].Build)
		assert.Equal(t, map[string]string{"FOO": "baz"}, rs[0].Env)
		assert.Equal(t, "B3", rs[1].Build)
		assert.Equal(t, map[string]string{"FOO": "baz"}, rs[1].Env)
		assert.Equal(t, "B3", rs[2].Build)
		assert.Equal(t, map[string]string{"FOO": "bar"}, rs[2].Env)
		assert.Equal(t, "B2", rs[3].Build)
		assert.Equal(t, map[string]string{"FOO": "bar"}, rs[3].Env)
		assert.Equal(t, "B1", rs[4].Build)
		assert.Equal(t, map[string]string{"FOO": "bar"}, rs[4].Env)
		assert.Equal(t, "B1", rs[5].Build)
		assert.Equal(t, map[string]string(nil), rs[5].Env)
	}

	// expects := types.Releases{
	//   {App: "app", Build: "BTEST3", Env: map[string]string{"APP": "app", "FOO": "baz"}},
	//   {App: "app", Build: "BTEST2", Env: map[string]string{"APP": "app", "FOO": "bar"}},
	//   {App: "app", Build: "BTEST", Env: map[string]string{"APP": "app", "FOO": "bar"}},
	// }

	// opts := types.ReleaseCreateOptions{
	//   Build: "BTEST",
	//   Env: map[string]string{
	//     "APP": "app",
	//     "FOO": "bar",
	//   },
	// }
	// var r *types.Release
	// r, err = p.ReleaseCreate("app", opts)
	// fmt.Printf("r = %+v\n", r)
	// assert.NoError(t, err)

	// opts.Build = "BTEST2"
	// r, err = p.ReleaseCreate("app", opts)
	// fmt.Printf("r = %+v\n", r)
	// assert.NoError(t, err)

	// opts.Build = "BTEST3"
	// opts.Env["FOO"] = "baz"
	// r, err = p.ReleaseCreate("app", opts)
	// fmt.Printf("r = %+v\n", r)
	// assert.NoError(t, err)

	// rels, err := p.ReleaseList("app")
	// assert.Len(t, rels, 3)

	// fmt.Printf("rels = %+v\n", rels)

	// for i := range rels {
	//   assert.Equal(t, expects[i].App, rels[i].App)
	//   assert.Equal(t, expects[i].Build, rels[i].Build)
	//   assert.Equal(t, expects[i].Env, rels[i].Env)
	//   assert.Equal(t, false, rels[i].Created.IsZero())
	//   assert.NotEqual(t, "", rels[i].Id)
	// }
}
