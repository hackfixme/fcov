package report

import (
	"testing"

	"github.com/friendlycaptcha/fcov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	cov := &types.Coverage{Files: map[string]map[types.FileBlock]*types.Stats{
		"pkg1/file1.go": {
			{Start: types.FileLocation{Line: 16, Col: 47},
				End: types.FileLocation{Line: 18, Col: 3},
			}: {NumStatements: 1, HitCount: 2},
			{Start: types.FileLocation{Line: 22, Col: 13},
				End: types.FileLocation{Line: 40, Col: 2},
			}: {NumStatements: 3, HitCount: 0},
		},
		"pkg2/file1.go": {
			{Start: types.FileLocation{Line: 10, Col: 21},
				End: types.FileLocation{Line: 22, Col: 5},
			}: {NumStatements: 2, HitCount: 4},
			{Start: types.FileLocation{Line: 24, Col: 18},
				End: types.FileLocation{Line: 30, Col: 6},
			}: {NumStatements: 1, HitCount: 1},
		},
		"pkg2/file2.go": {
			{Start: types.FileLocation{Line: 10, Col: 16},
				End: types.FileLocation{Line: 18, Col: 5},
			}: {NumStatements: 3, HitCount: 0},
			{Start: types.FileLocation{Line: 22, Col: 15},
				End: types.FileLocation{Line: 28, Col: 3},
			}: {NumStatements: 2, HitCount: 0},
		},
	}}
	rep := Create(cov)
	assert.NotNil(t, rep)

	assert.Equal(t, 12, rep.NumStatements)
	assert.Equal(t, 4, rep.HitCount)
	assert.Equal(t, 0.3333333333333333, rep.Coverage)

	require.Len(t, rep.Packages, 2)

	require.Contains(t, rep.Packages, "pkg1")
	pkg1 := rep.Packages["pkg1"]
	assert.Equal(t, 4, pkg1.NumStatements)
	assert.Equal(t, 1, pkg1.HitCount)
	assert.Equal(t, 0.25, pkg1.Coverage)

	require.Len(t, pkg1.Files, 1)
	require.Contains(t, pkg1.Files, "file1.go")
	p1f1 := pkg1.Files["file1.go"]
	assert.Equal(t, 4, p1f1.NumStatements)
	assert.Equal(t, 1, p1f1.HitCount)
	assert.Equal(t, 0.25, p1f1.Coverage)

	require.Contains(t, rep.Packages, "pkg2")
	pkg2 := rep.Packages["pkg2"]
	assert.Equal(t, 8, pkg2.NumStatements)
	assert.Equal(t, 3, pkg2.HitCount)
	assert.Equal(t, 0.375, pkg2.Coverage)

	require.Len(t, pkg2.Files, 2)
	require.Contains(t, pkg2.Files, "file1.go")
	p2f1 := pkg2.Files["file1.go"]
	assert.Equal(t, 3, p2f1.NumStatements)
	assert.Equal(t, 3, p2f1.HitCount)
	assert.Equal(t, 1.0, p2f1.Coverage)

	require.Contains(t, pkg2.Files, "file2.go")
	p2f2 := pkg2.Files["file2.go"]
	assert.Equal(t, 5, p2f2.NumStatements)
	assert.Equal(t, 0, p2f2.HitCount)
	assert.Equal(t, 0.0, p2f2.Coverage)
}
