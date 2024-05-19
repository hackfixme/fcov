# fcov

fcov is a tool that analyzes code coverage files, and generates reports in
various formats. It currently supports Go coverage files, and text and Markdown
formats.

It can be used to quickly visualize coverage on the command line, or made part
of a CI pipeline to generate coverage reports that can be posted as comments
in pull requests.


## Installation

You can download a package for Windows, macOS and Linux from the
[releases page](https://github.com/FriendlyCaptcha/fcov/releases). Extract the
archive and move the binary to a directory in your system's `$PATH`.

Alternatively, you can build your own binary with these instructions.

First, ensure that you have installed
[Git](https://github.com/git-guides/install-git) and the latest
[Go toolchain](https://golang.org/doc/install) (at least version 1.21).

Then run:

```sh
go install github.com/friendlycaptcha/fcov@latest
```


## Usage

The tool has a command-line interface.

### Report

The `report` command reads one or more coverage files, and generates reports that can
be written to stdout, or one or more files.


#### Options

- `--filter`: accepts one or more glob patterns in
  [`gitignore` format](https://git-scm.com/docs/gitignore) for specifying
  package or file paths to include or exclude from coverage processing *and*
  the generated report.

- `--filter-output`: accepts one or more glob patterns in
  [`gitignore` format](https://git-scm.com/docs/gitignore) for specifying
  package or file paths to include or exclude *only* from the generated
  report.

  That is, if a path is excluded with `--filter`, then the calculated coverage
  percentage might be affected, and the path will also not be displayed in the
  report. However, if the same path is only excluded with
  `--filter-output` but not `--filter`, then the coverage percentage will not be
  affected, and the path will only be hidden from the report.

  This behavior is useful if you want to show the correct global coverage,
  but only create a report of specific packages or files. For example, to
  only show files and packages changed in a pull request.

- `--nest-files`, `--no-nest-files`: enable or disable file nesting
  under packages. This is useful for removing the repetition of the package path
  from the files that belong to that package.  
  Default: `--nest-files`

- `--output` / `-o`: Write the report to stdout, and/or one or more files.
  More than one value can be provided, separated by comma. If a value is either
  `'txt'` or `'md'`, the report will be written to stdout in text or Markdown
  format, respectively. If a value is in the form of a filename, e.g.
  `'report.md'`, then it will be written to a file with the format inferred from
  the extension.  
  Default: `'txt'`

- `--thresholds`: Lower and upper thresholds separated by comma used to change
  the output depending on the coverage percentage. For example, this is used by
  the Markdown format to change the color of the badge and coverage indicators.  
  Default: `'50,75'`

- `--trim-package-prefix`: Value to trim from the file path prefix in the
  output. This is useful for removing long and common package names, to keep the
  output tidier.


#### Examples

- Process a single coverage file using default options:
  ```sh
  $ fcov report coverage.txt
  ```

  This outputs a report in text format with files nested under each package:
  ```
  github.com/friendlycaptcha/fcov/app      100.00%
      app.go                               100.00%
      options.go                           100.00%
  github.com/friendlycaptcha/fcov/app/cli   83.33%
      cli.go                                91.67%
      report.go                             81.25%
  github.com/friendlycaptcha/fcov/cmd/fcov   0.00%
      main.go                                0.00%
  github.com/friendlycaptcha/fcov/parse    100.00%
      go.go                                100.00%
  github.com/friendlycaptcha/fcov/report   100.00%
      render.go                            100.00%
      report.go                            100.00%
  github.com/friendlycaptcha/fcov/types     80.00%
      types.go                              80.00%
  
  Total Coverage: 94.16%
  ```

- Process multiple coverage files using default options:
  ```sh
  $ fcov report coverage1.txt coverage2.txt ...
  ```
  
  Or using shell globbing:
  ```sh
  $ fcov report coverage*.txt
  ```

- Exclude generated Go files based on their extension:
  ```sh
  $ fcov report --filter '*[._]gen.go,*.pb.go' coverage.txt
  ```

  Make sure that the `--filter` value is quoted to prevent it from being
  interpreted by the shell.

  Note that multiple patterns can be separated with a comma, and the use of the
  range notation to match both `*.gen.go` and `*_gen.go` files. See the
  [`gitignore` format documentation](https://git-scm.com/docs/gitignore)
  for other syntax examples.

- Exclude all files except the `fcov/report` package:
  ```sh
  $ fcov report --filter '*,!fcov/report' coverage.txt
  github.com/friendlycaptcha/fcov/report 100.00%
      render.go                          100.00%
      report.go                          100.00%
  
  Total Coverage: 100.00%
  ```
    
  The `!` prefix can be used to negate a pattern. 

- Exclude all files only from the report, except the `fcov/report` package:
  ```sh
  $ fcov report --filter-output '*,!fcov/report' coverage.txt
  github.com/friendlycaptcha/fcov/report 100.00%
      render.go                          100.00%
      report.go                          100.00%
  
  Total Coverage: 94.16%
  ```
  
  Note the difference in Total Coverage from the `--filter` example above.
  It is lower since all project files were used for calculating it, but only
  files in the `fcov/report` package are shown.

- Disable file nesting below packages:
  ```sh
  $ fcov report --filter-output '*,!fcov/report' --no-nest-files coverage.txt
  github.com/friendlycaptcha/fcov/report           100.00%
  github.com/friendlycaptcha/fcov/report/render.go 100.00%
  github.com/friendlycaptcha/fcov/report/report.go 100.00%
  
  Total Coverage: 94.16%
  ```

- Write the report to stdout in text format, and write it to a `report.md`
  file in Markdown format:
  ```sh
  $ fcov report --output 'txt,report.md' coverage.txt
  ```

- Output the report in Markdown format to stdout:
  ```sh
  $ fcov report --output md coverage.txt
  ```

  Here's what it looks like rendered:

  <hr>

  ![Total Coverage](https://img.shields.io/badge/Total%20Coverage-94.16%25-success?style=flat)
  
  | Package                                                                                                                                                                                 | Coverage |
  | :------                                                                                                                                                                                 | -------: |
  | <details><summary>`github.com/friendlycaptcha/fcov/app`</summary><table><tr><td>`app.go`</td><td>100.00%</td></tr><tr><td>`options.go`</td><td>100.00%</td></tr></table></details>      |  100.00% |
  | <details><summary>`github.com/friendlycaptcha/fcov/app/cli`</summary><table><tr><td>`cli.go`</td><td>91.67%</td></tr><tr><td>`report.go`</td><td>81.25%</td></tr></table></details>     |   83.33% |
  | <details><summary>`github.com/friendlycaptcha/fcov/cmd/fcov`</summary><table><tr><td>`main.go`</td><td>0.00%</td></tr></table></details>                                                |    0.00% |
  | <details><summary>`github.com/friendlycaptcha/fcov/parse`</summary><table><tr><td>`go.go`</td><td>100.00%</td></tr></table></details>                                                   |  100.00% |
  | <details><summary>`github.com/friendlycaptcha/fcov/report`</summary><table><tr><td>`render.go`</td><td>100.00%</td></tr><tr><td>`report.go`</td><td>100.00%</td></tr></table></details> |  100.00% |
  | <details><summary>`github.com/friendlycaptcha/fcov/types`</summary><table><tr><td>`types.go`</td><td>80.00%</td></tr></table></details>                                                 |   80.00% |

  <hr>

  View the source of this README file to see the raw Markdown.

  You can use the `--no-nest-files` option to disable the collapsible element,
  and show each file on its own line:
  ```sh
  $ fcov report --output md --no-nest-files coverage.txt
  ```

  <hr>

  ![Total Coverage](https://img.shields.io/badge/Total%20Coverage-94.16%25-success?style=flat)
  
  | Package                                             | Coverage |
  | :------                                             | -------: |
  | `github.com/friendlycaptcha/fcov/app`               |  100.00% |
  | `github.com/friendlycaptcha/fcov/app/app.go`        |  100.00% |
  | `github.com/friendlycaptcha/fcov/app/options.go`    |  100.00% |
  | `github.com/friendlycaptcha/fcov/app/cli`           |   83.33% |
  | `github.com/friendlycaptcha/fcov/app/cli/cli.go`    |   91.67% |
  | `github.com/friendlycaptcha/fcov/app/cli/report.go` |   81.25% |
  | `github.com/friendlycaptcha/fcov/cmd/fcov`          |    0.00% |
  | `github.com/friendlycaptcha/fcov/cmd/fcov/main.go`  |    0.00% |
  | `github.com/friendlycaptcha/fcov/parse`             |  100.00% |
  | `github.com/friendlycaptcha/fcov/parse/go.go`       |  100.00% |
  | `github.com/friendlycaptcha/fcov/report`            |  100.00% |
  | `github.com/friendlycaptcha/fcov/report/render.go`  |  100.00% |
  | `github.com/friendlycaptcha/fcov/report/report.go`  |  100.00% |
  | `github.com/friendlycaptcha/fcov/types`             |   80.00% |
  | `github.com/friendlycaptcha/fcov/types/types.go`    |   80.00% |

  <hr>

- Use different coverage thresholds to change the color of the badge in the
  Markdown report. With the default thresholds of `'50,75'`, a total coverage
  value below 50% will generate a red badge, between 50% and 75% a yellow badge,
  and above 75% a green badge. To set the lower threshold to 40% and upper to
  60% run:
  ```sh
  $ fcov report --output md --thresholds '40,60' coverage.txt
  ```

- Trim a common package prefix:
  ```sh
  $ fcov report --trim-package-prefix github.com/friendlycaptcha/ coverage.txt
  fcov/app       100.00%
      app.go     100.00%
      options.go 100.00%
  fcov/app/cli    83.33%
      cli.go      91.67%
      report.go   81.25%
  fcov/cmd/fcov    0.00%
      main.go      0.00%
  fcov/parse     100.00%
      go.go      100.00%
  fcov/report    100.00%
      render.go  100.00%
      report.go  100.00%
  fcov/types      80.00%
      types.go    80.00%
  
  Total Coverage: 94.16%
  ```


## License

[MIT](LICENSE)
