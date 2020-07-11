#!/usr/bin/env bash

set -x

GIT_SHA=$(git rev-parse --short HEAD || echo "GitNotFound")
VERSION=$(git describe --tags || echo "VersionNotFound")

if [[ -n "$FAILPOINTS" ]]; then
	GIT_SHA="$GIT_SHA"-FAILPOINTS
fi

GO_LDFLAGS="$GO_LDFLAGS -X github.com/mayongze/joss-cli/command.GitSHA=${GIT_SHA} -X github.com/mayongze/joss-cli/command.Version=${VERSION}"

OUTPUT=output
if [[ -n "${BINDIR}" ]]; then OUTPUT="${BINDIR}"; fi
rm -rf ./$OUTPUT && mkdir $OUTPUT

GOOS=windows GOARCH=386    go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-windows-x86.exe joss-cli.go
GOOS=windows GOARCH=amd64  go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-windows-x64.exe joss-cli.go
GOOS=darwin  GOARCH=amd64  go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-darwin-x64 joss-cli.go
GOOS=linux   GOARCH=amd64  go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-linux-x64 joss-cli.go
GOOS=linux   GOARCH=386    go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-linux-x86 joss-cli.go
GOOS=linux   GOARCH=arm    go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-linux-arm joss-cli.go
GOOS=linux   GOARCH=mipsle go build $GO_BUILD_FLAGS -ldflags "$GO_LDFLAGS" -o $OUTPUT/joss-linux-mipsle joss-cli.go

# build package
if [[ ! -n "${VERSION}" ]]; then VERSION=`git describe --tags`; fi
# shellcheck disable=SC2164
pushd ./${OUTPUT}
tar -czvf joss-${VERSION}-windows-x86.tar.gz joss-windows-x86.exe
tar -czvf joss-${VERSION}-windows-x64.tar.gz joss-windows-x64.exe
tar -czvf joss-${VERSION}-darwin-x64.tar.gz joss-darwin-x64
tar -czvf joss-${VERSION}-linux-x64.tar.gz joss-linux-x64
tar -czvf joss-${VERSION}-linux-x86.tar.gz joss-linux-x86
tar -czvf joss-${VERSION}-linux-arm.tar.gz joss-linux-arm
tar -czvf joss-${VERSION}-linux-mipsle.tar.gz joss-linux-mipsle

# shellcheck disable=SC2164
popd
