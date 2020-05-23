#!/bin/bash

OUTPUT=output
rm -rf ./$OUTPUT && mkdir $OUTPUT

GOOS=windows GOARCH=386    go build -o $OUTPUT/joss-windows-x86.exe joss-cli.go
GOOS=windows GOARCH=amd64  go build -o $OUTPUT/joss-windows-x64.exe joss-cli.go
GOOS=darwin  GOARCH=amd64  go build -o $OUTPUT/joss-darwin-x64 joss-cli.go
GOOS=linux   GOARCH=amd64  go build -o $OUTPUT/joss_linux_x64 joss-cli.go
GOOS=linux   GOARCH=386    go build -o $OUTPUT/joss_linux_x86 joss-cli.go
GOOS=linux   GOARCH=arm    go build -o $OUTPUT/joss-linux-arm joss-cli.go