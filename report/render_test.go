package report

import (
	"testing"
	"text/template"

	gitignore "github.com/sabhiram/go-gitignore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.hackfix.me/fcov/types"
)

func TestReportRender(t *testing.T) {
	t.Parallel()

	report := &Report{
		Stats: types.Stats{
			Coverage: 0.84899,
		},
		Packages: map[string]*Package{
			"path/pkg1": {
				Stats: types.Stats{
					Coverage: 0.7512,
				},
				Name: "path/pkg1",
				Files: map[string]*File{
					"file1.go": {
						Stats: types.Stats{
							Coverage: 0.3542,
						},
						Name:    "file1.go",
						Package: "path/pkg1",
					},
					"file2.go": {
						Stats: types.Stats{
							Coverage: 0.9747,
						},
						Name:    "file2.go",
						Package: "path/pkg1",
					},
				},
			},
			"path/pkg2": {
				Stats: types.Stats{
					Coverage: 0.6486,
				},
				Name: "path/pkg2",
				Files: map[string]*File{
					"file3.go": {
						Stats: types.Stats{
							Coverage: 0.4781,
						},
						Name:    "file3.go",
						Package: "path/pkg2",
					},
				},
			},
		},
	}

	var tests = []struct {
		name              string
		format            Format
		nestFiles         bool
		filter            *gitignore.GitIgnore
		trimPackagePrefix string
		want              string
	}{
		{
			name:              "txt_nest_nofilter_notrim",
			format:            Text,
			nestFiles:         true,
			filter:            gitignore.CompileIgnoreLines(""),
			trimPackagePrefix: "",
			want: "path/pkg1    75.12% \n" +
				"    file1.go 35.42% \n" +
				"    file2.go 97.47% \n" +
				"path/pkg2    64.86% \n" +
				"    file3.go 47.81% \n\n" +
				"Total Coverage: 84.90%",
		},
		{
			name:              "txt_nonest_filter_trim",
			format:            Text,
			nestFiles:         false,
			filter:            gitignore.CompileIgnoreLines("*/pkg1"),
			trimPackagePrefix: "path/",
			// FIXME: The package prefix shouldn't be rendered.
			want: "\x00pkg2          64.86% \n" +
				"pkg2/file3.go 47.81% \n\n" +
				"Total Coverage: 84.90%",
		},
		{
			name:              "md_nest_nofilter_trim",
			format:            Markdown,
			nestFiles:         true,
			filter:            gitignore.CompileIgnoreLines(""),
			trimPackagePrefix: "path/",
			want: "![Total Coverage](https://img.shields.io/badge/Total%20Coverage-84.90%25-yellow?style=flat)\n\n" +
				"| Package                                                                                                                                           | Coverage |\n" +
				"| :------                                                                                                                                           | -------: |\n" +
				"| <details><summary>`pkg1`</summary><table><tr><td>`file1.go`</td><td>35.42%</td></tr><tr><td>`file2.go`</td><td>97.47%</td></tr></table></details> |   75.12% |\n" +
				"| <details><summary>`pkg2`</summary><table><tr><td>`file3.go`</td><td>47.81%</td></tr></table></details>                                            |   64.86% |",
		},
		{
			name:              "md_nonest_filter_notrim",
			format:            Markdown,
			nestFiles:         false,
			filter:            gitignore.CompileIgnoreLines("file3.go"),
			trimPackagePrefix: "",

			want: "![Total Coverage](https://img.shields.io/badge/Total%20Coverage-84.90%25-yellow?style=flat)\n\n" +
				"| Package              | Coverage |\n" +
				"| :------              | -------: |\n" +
				"| `path/pkg1`          |   75.12% |\n" +
				"| `path/pkg1/file1.go` |   35.42% |\n" +
				"| `path/pkg1/file2.go` |   97.47% |\n" +
				"| `path/pkg2`          |   64.86% |",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := report.Render(tt.format, tt.nestFiles, tt.filter, 70, 90, tt.trimPackagePrefix)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		r := &Report{}
		assert.Equal(t, "", r.Render(Text, false, nil, 70, 90, ""))
	})
}

func TestRenderMarkdown(t *testing.T) {
	sum := [][]string{
		{string(pkgMarker) + "pkg1", "90.00%"},
		{"file1.go", "85.00%"},
		{"file2.go", "70.00%"},
		{string(pkgMarker) + "pkg2", "15.00%"},
		{"file3.go", "10.00%"},
	}
	var data [][]string

	renderMarkdown(sum, &data)

	require.Equal(t, 5, len(data))
	assert.Equal(t, "`pkg1`", data[0][0])
	assert.Equal(t, "90.00%", data[0][1])
	assert.Equal(t, "`file1.go`", data[1][0])
	assert.Equal(t, "85.00%", data[1][1])
	assert.Equal(t, "`file2.go`", data[2][0])
	assert.Equal(t, "70.00%", data[2][1])
	assert.Equal(t, "`pkg2`", data[3][0])
	assert.Equal(t, "15.00%", data[3][1])
	assert.Equal(t, "`file3.go`", data[4][0])
	assert.Equal(t, "10.00%", data[4][1])
}

func TestRenderMarkdownNested(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		t.Parallel()
		sum := [][]string{
			{string(pkgMarker) + "pkg1", "90.00%"},
			{"file1.go", "85.00%"},
			{"file2.go", "70.00%"},
			{string(pkgMarker) + "pkg2", "15.00%"},
			{"file3.go", "10.00%"},
		}
		var data [][]string

		renderMarkdownNested(sum, &data)

		require.Equal(t, 2, len(data))
		assert.Equal(t, "<details><summary>`pkg1`</summary><table>"+
			"<tr><td>`file1.go`</td><td>85.00%</td></tr>"+
			"<tr><td>`file2.go`</td><td>70.00%</td></tr></table></details>",
			data[0][0])
		assert.Equal(t, "90.00%", data[0][1])
		assert.Equal(t, "<details><summary>`pkg2`</summary><table>"+
			"<tr><td>`file3.go`</td><td>10.00%</td></tr></table></details>",
			data[1][0])
		assert.Equal(t, "15.00%", data[1][1])
	})

	t.Run("err/panic_template", func(t *testing.T) {
		t.Parallel()
		// Not using assert.PanicsWithError since it doesn't report messages for
		// complex errors correctly.
		// See https://github.com/stretchr/testify/issues/1399
		defer func() {
			err := recover()
			require.IsType(t, template.ExecError{}, err)
			tmplErr := err.(template.ExecError)
			assert.Equal(t, "table", tmplErr.Name)
			require.Error(t, tmplErr.Err)
			assert.Contains(t, tmplErr.Err.Error(), "reflect: slice index out of range")
		}()

		sum := [][]string{
			{string(pkgMarker) + "pkg1", "90.00%"},
			// The file line is missing the coverage at index 1, which should
			// fail during templating.
			{"file1.go"},
		}
		var data [][]string

		renderMarkdownNested(sum, &data)

		t.Errorf("did not panic")
	})
}

func TestRenderTextNested(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		name string
		sum  [][]string
		want [][]string
	}{
		{
			name: "no_pkg_line",
			sum: [][]string{
				{"file1.go", "6.32%"},
				{"file2.go", "97.06%"},
			},
			want: [][]string{
				{"    file1.go", "6.32%"},
				{"    file2.go", "97.06%"},
			},
		},
		{
			name: "with_pkg_line",
			sum: [][]string{
				{string(pkgMarker) + "pkg1", "30.23%"},
				{"file1.go", "6.32%"},
			},
			want: [][]string{
				{"pkg1", "30.23%"},
				{"    file1.go", "6.32%"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := [][]string{}
			renderTextNested(tt.sum, &data)
			assert.Equal(t, tt.want, data)
		})
	}
}

func TestFormatFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Format
	}{
		{"txt", Text},
		{"md", Markdown},
		{"unknown", ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := FormatFromString(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateBadgeURL(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		cov            float64
		lowerThreshold float64
		upperThreshold float64
		want           string
	}{
		{80, 70, 90,
			"https://img.shields.io/badge/Total%20Coverage-80.00%25-yellow?style=flat"},
		{60, 70, 90,
			"https://img.shields.io/badge/Total%20Coverage-60.00%25-critical?style=flat"},
		{90, 70, 90,
			"https://img.shields.io/badge/Total%20Coverage-90.00%25-success?style=flat"},
	}
	for _, tt := range tests {
		got := generateBadgeURL(tt.cov, tt.lowerThreshold, tt.upperThreshold)
		assert.Equal(t, tt.want, got)
	}
}
