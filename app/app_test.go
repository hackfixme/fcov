package app

import (
	"bufio"
	"bytes"
	"errors"
	"log/slog"
	"os"
	"sync"
	"testing"

	actx "github.com/friendlycaptcha/fcov/app/context"

	"github.com/mandelsoft/vfs/pkg/memoryfs"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("ok/help", func(t *testing.T) {
		t.Parallel()

		var stdin, stdout, stderr bytes.Buffer
		stderrW := bufio.NewWriter(&stderr)
		stdoutW := bufio.NewWriter(&stdout)
		logger := slog.New(slog.NewTextHandler(stderrW, nil))

		assert.PanicsWithError(t, "test interrupt", func() {
			New(
				WithArgs([]string{"--help"}),
				WithEnv(&mockEnv{env: map[string]string{}}),
				WithLogger(logger),
				WithExit(func(int) {
					panic(errors.New("test interrupt")) // Panic to fake "exit"
				}),
				WithFS(memoryfs.New()),
				WithFDs(bufio.NewReader(&stdin), stdoutW, stderrW),
			)
		})

		stdoutW.Flush()
		stderrW.Flush()
		assert.Contains(t, stdout.String(), "Usage: fcov <command>")
		assert.Equal(t, "", stderr.String())
	})

	t.Run("ok/report_stdout", func(t *testing.T) {
		t.Parallel()

		var stdin, stdout, stderr bytes.Buffer
		stderrW := bufio.NewWriter(&stderr)
		stdoutW := bufio.NewWriter(&stdout)
		logger := slog.New(slog.NewTextHandler(stderrW, nil))

		covData, err := os.ReadFile("testdata/coverage_ok_atomic.txt")
		require.NoError(t, err)
		memfs := memoryfs.New()
		err = vfs.WriteFile(memfs, "/coverage_ok_atomic.txt", covData, 0644)
		require.NoError(t, err)

		New(
			WithArgs([]string{"report", "/coverage_ok_atomic.txt"}),
			WithLogger(logger),
			WithExit(func(int) {
				panic("test interrupt") // Panic to fake "exit"
			}),
			WithFS(memfs),
			WithFDs(bufio.NewReader(&stdin), stdoutW, stderrW),
		).Run()

		stdoutW.Flush()
		stderrW.Flush()
		// Not using a raw string literal since tablewriter adds a trailing
		// space to each line, which I don't know how to disable, and my editor
		// is configured to remove trailing whitespace...
		expOut := "pkg1         72.41% \n" +
			"    file1.go 60.00% \n" +
			"    file2.go 78.95% \n" +
			"pkg2         37.25% \n" +
			"    file1.go  2.50% \n" +
			"    file2.go 59.68% \n\n" +
			"Total Coverage: 45.04%\n"
		assert.Equal(t, expOut, stdout.String())
		assert.Equal(t, "", stderr.String())
	})

	t.Run("ok/report_file", func(t *testing.T) {
		t.Parallel()

		var stdin, stdout, stderr bytes.Buffer
		stderrW := bufio.NewWriter(&stderr)
		stdoutW := bufio.NewWriter(&stdout)
		logger := slog.New(slog.NewTextHandler(stderrW, nil))

		covData, err := os.ReadFile("testdata/coverage_ok_atomic.txt")
		require.NoError(t, err)
		memfs := memoryfs.New()
		err = vfs.WriteFile(memfs, "/coverage_ok_atomic.txt", covData, 0644)
		require.NoError(t, err)

		New(
			WithArgs([]string{"report", "--output=/report.txt,/report.md", "/coverage_ok_atomic.txt"}),
			WithLogger(logger),
			WithExit(func(int) {
				panic("test interrupt") // Panic to fake "exit"
			}),
			WithFS(memfs),
			WithFDs(bufio.NewReader(&stdin), stdoutW, stderrW),
		).Run()

		stdoutW.Flush()
		stderrW.Flush()
		assert.Equal(t, "", stdout.String())
		assert.Equal(t, "", stderr.String())

		reportTxt, err := vfs.ReadFile(memfs, "/report.txt")
		require.NoError(t, err)
		expReportTxt := "pkg1         72.41% \n" +
			"    file1.go 60.00% \n" +
			"    file2.go 78.95% \n" +
			"pkg2         37.25% \n" +
			"    file1.go  2.50% \n" +
			"    file2.go 59.68% \n\n" +
			"Total Coverage: 45.04%"
		assert.Equal(t, expReportTxt, string(reportTxt))

		reportMd, err := vfs.ReadFile(memfs, "/report.md")
		require.NoError(t, err)
		expReportMd := `![Total Coverage](https://img.shields.io/badge/Total%20Coverage-45.04%25-critical?style=flat)

| Package                                                                                                                                           | Coverage |
| :------                                                                                                                                           | -------: |
| <details><summary>` + "`pkg1`" + `</summary><table><tr><td>` + "`file1.go`" + `</td><td>60.00%</td></tr><tr><td>` + "`file2.go`" + `</td><td>78.95%</td></tr></table></details> |   72.41% |
| <details><summary>` + "`pkg2`" + `</summary><table><tr><td>` + "`file1.go`" + `</td><td>2.50%</td></tr><tr><td>` + "`file2.go`" + `</td><td>59.68%</td></tr></table></details>  |   37.25% |`
		assert.Equal(t, expReportMd, string(reportMd))
	})

	t.Run("err/cli_args", func(t *testing.T) {
		t.Parallel()

		var stdin, stdout, stderr bytes.Buffer
		stderrW := bufio.NewWriter(&stderr)
		stdoutW := bufio.NewWriter(&stdout)
		logger := slog.New(slog.NewTextHandler(stderrW, nil))

		assert.PanicsWithError(t, "test interrupt", func() {
			New(
				WithArgs([]string{"missingcommand"}),
				WithLogger(logger),
				WithExit(func(int) {
					panic(errors.New("test interrupt")) // Panic to fake "exit"
				}),
				WithFS(memoryfs.New()),
				WithFDs(bufio.NewReader(&stdin), stdoutW, stderrW),
			)
		})

		stdoutW.Flush()
		stderrW.Flush()
		assert.Equal(t, "", stdout.String())
		assert.Contains(t, stderr.String(), "unexpected argument missingcommand")
	})
}

type mockEnv struct {
	mx  sync.RWMutex
	env map[string]string
}

var _ actx.Environment = &mockEnv{}

func (me *mockEnv) Get(key string) string {
	me.mx.RLock()
	defer me.mx.RUnlock()
	return me.env[key]
}

func (me *mockEnv) Set(key, val string) error {
	me.mx.Lock()
	defer me.mx.Unlock()
	me.env[key] = val
	return nil
}
