package cli

import (
	"strings"
	"testing"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: Add more test cases.
func TestCreateOutputFilterFromFile(t *testing.T) {
	t.Parallel()

	fs := memoryfs.New()
	f, err := vfs.TempFile(fs, "/", "")
	require.NoError(t, err)
	defer func() {
		cerr := f.Close()
		assert.NoError(t, cerr)
		cerr = vfs.Cleanup(fs)
		assert.NoError(t, cerr)
	}()

	filterFiles := []string{
		"pkg1/file1.go",
		"pkg1/file2.go",
		"pkg2/file1.go",
		"pkg2/file2.go",
		"pkg2/file1_test.go",
	}

	_, err = f.WriteString(strings.Join(filterFiles, "\n"))

	// Prepare for reading
	_, err = f.Seek(0, 0)
	require.NoError(t, err)

	filterOut, err := createOutputFilterFromFile(f)
	require.NoError(t, err)

	assert.Equal(t, []string{"*", "!pkg1/file1.go", "!pkg1/file2.go", "!pkg2/"}, filterOut)
}
