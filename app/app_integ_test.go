package app

import (
	"os"
	"testing"
	"time"

	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApp(t *testing.T) {
	t.Parallel()

	t.Run("ok/report_stdout", func(t *testing.T) {
		t.Parallel()

		tctx, cancel, h := newTestContext(t, 5*time.Second)
		defer cancel()
		app, err := newTestApp(tctx)
		h(assert.NoError(t, err))

		covData, err := os.ReadFile("testdata/coverage_ok_atomic.txt")
		require.NoError(t, err)
		err = vfs.WriteFile(app.ctx.FS, "/coverage_ok_atomic.txt", covData, 0o644)
		require.NoError(t, err)

		err = app.Run("report", "/coverage_ok_atomic.txt")
		require.NoError(t, err)

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

		h(assert.Equal(t, expOut, app.stdout.String()))
		h(assert.Equal(t, "", app.stderr.String()))
	})

	t.Run("ok/report_file", func(t *testing.T) {
		t.Parallel()

		tctx, cancel, h := newTestContext(t, 5*time.Second)
		defer cancel()
		app, err := newTestApp(tctx)
		h(assert.NoError(t, err))

		covData, err := os.ReadFile("testdata/coverage_ok_atomic.txt")
		require.NoError(t, err)
		err = vfs.WriteFile(app.ctx.FS, "/coverage_ok_atomic.txt", covData, 0o644)
		require.NoError(t, err)

		err = app.Run("report", "--output=/report.txt,/report.md", "/coverage_ok_atomic.txt")
		require.NoError(t, err)

		reportTxt, err := vfs.ReadFile(app.ctx.FS, "/report.txt")
		require.NoError(t, err)
		expReportTxt := "pkg1         72.41% \n" +
			"    file1.go 60.00% \n" +
			"    file2.go 78.95% \n" +
			"pkg2         37.25% \n" +
			"    file1.go  2.50% \n" +
			"    file2.go 59.68% \n\n" +
			"Total Coverage: 45.04%"
		h(assert.Equal(t, expReportTxt, string(reportTxt)))

		reportMd, err := vfs.ReadFile(app.ctx.FS, "/report.md")
		require.NoError(t, err)
		expReportMd := `![Total Coverage](https://img.shields.io/badge/Total%20Coverage-45.04%25-critical?style=flat)

| Package                                                                                                                                           | Coverage |
| :------                                                                                                                                           | -------: |
| <details><summary>` + "`pkg1`" + `</summary><table><tr><td>` + "`file1.go`" + `</td><td>60.00%</td></tr><tr><td>` + "`file2.go`" + `</td><td>78.95%</td></tr></table></details> |   72.41% |
| <details><summary>` + "`pkg2`" + `</summary><table><tr><td>` + "`file1.go`" + `</td><td>2.50%</td></tr><tr><td>` + "`file2.go`" + `</td><td>59.68%</td></tr></table></details>  |   37.25% |`
		h(assert.Equal(t, expReportMd, string(reportMd)))
	})
}
