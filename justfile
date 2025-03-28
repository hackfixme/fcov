version := "1"

default:
  just --list


build *ARGS:
  ./release/build.sh '{{ARGS}}'


clean:
  rm -rf ./dist ./golangci-lint*.txt


lint report="":
  #!/usr/bin/env sh
  if [ -z '{{report}}' ]; then
    golangci-lint run --timeout 5m --out-format=tab --new-from-rev=fa1e6fe876 ./...
    exit $?
  fi

  _report_id="$(date '+%Y%m%d')-$(git describe --tags --abbrev=10 --always)"
  golangci-lint run --timeout 5m --out-format=tab --issues-exit-code=0 ./... | \
    tee "golangci-lint-${_report_id}.txt" | \
      awk 'NF {if ($2 == "revive") print $2 ":" $3; else print $2}' \
      | sort | uniq -c | sort -nr \
      | tee "golangci-lint-summary-${_report_id}.txt"


test target="..." *ARGS="":
  go test -v -race -count=1 -failfast {{ARGS}} ./{{target}}