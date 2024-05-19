package parse

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friendlycaptcha/fcov/types"
)

func TestGo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		covFile  string
		filter   []string
		expErr   string
		expFiles map[string]map[string][2]int
	}{
		{
			name:    "ok/empty",
			covFile: "coverage_ok_empty.txt",
		},
		{
			name:    "ok/atomic/no_filter",
			covFile: "coverage_ok_atomic.txt",
			expFiles: map[string]map[string][2]int{
				"pkg1/file1.go": {
					"16.47,18.3": {1, 0},
					"33.55,36.3": {2, 0},
				},
				"pkg1/file2.go": {
					"32.42,35.3":  {2, 0},
					"38.13,44.12": {4, 1},
				},
				"pkg2/file1.go": {
					"39.45,41.35": {2, 0},
					"62.3,64.38":  {3, 0},
				},
				"pkg2/file2.go": {
					"18.74,33.76": {7, 0},
					"70.98,73.64": {2, 221},
					"88.41,94.54": {5, 4438130},
				},
			},
		},
		{
			name:    "ok/atomic/filter_all",
			covFile: "coverage_ok_atomic.txt",
			filter:  []string{"*"},
		},
		{
			name:    "ok/atomic/filter_two",
			covFile: "coverage_ok_atomic.txt",
			filter:  []string{"*", "!*/file1.go"},
			expFiles: map[string]map[string][2]int{
				"pkg1/file1.go": {
					"32.55,33.55": {1, 0},
				},
				"pkg2/file1.go": {
					"41.35,42.21": {1, 1},
				},
			},
		},
		{
			name:    "err/parse_line",
			covFile: "coverage_err_parse_line.txt",
			expErr:  "failed parsing line 'pkg1/file1.go|16.47,18.3 1 0': wrong format",
		},
		{
			name:    "err/parse_block",
			covFile: "coverage_err_parse_block.txt",
			expErr:  "failed parsing line 'pkg1/file1.go:abcd.47,18.3 1 0': expected integer",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			covData, err := os.ReadFile(filepath.Join("testdata", tc.covFile))
			require.NoError(t, err)

			cov := types.NewCoverage()
			err = Go(bytes.NewReader(covData), cov,
				gitignore.CompileIgnoreLines(tc.filter...))
			if tc.expErr != "" {
				assert.EqualError(t, err, tc.expErr)
				return
			}
			require.NoError(t, err)

			require.Len(t, cov.Files, len(tc.expFiles))

			for expFname, expBlocks := range tc.expFiles {
				blocks, ok := cov.Files[expFname]
				if !assert.Truef(t, ok, "file not found in coverage: '%s'", expFname) {
					continue
				}

				for expBlockStr, expStats := range expBlocks {
					expFB := types.FileBlock{}
					err = expFB.UnmarshalText([]byte(expBlockStr))
					assert.NoErrorf(t, err,
						"file '%s': failed unmarshalling FileBlock data: %s",
						expFname, expBlockStr)

					block, ok := blocks[expFB]
					if !assert.Truef(t, ok, "file '%s': block not found: '%s'", expFname, expBlockStr) {
						continue
					}
					assert.Equalf(t, expStats[0], block.NumStatements,
						"file '%s': block '%s': unexpected number of statements",
						expFname, expBlockStr)
					assert.Equalf(t, expStats[1], block.HitCount,
						"file '%s': block '%s': unexpected hit count",
						expFname, expBlockStr)
				}
			}
		})
	}

	t.Run("err/scanner_read", func(t *testing.T) {
		t.Parallel()
		err := Go(mockReader{}, nil, nil)
		require.EqualError(t, err, "failed scanning input: read error")
	})
}

type mockReader struct{}

func (r mockReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}
