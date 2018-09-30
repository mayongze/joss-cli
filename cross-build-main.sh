GOOS=windows GOARCH=386    go build -o $GOPATH/joss-windows-x86.exe   main.go
GOOS=windows GOARCH=amd64  go build -o $GOPATH/joss-windows-x64.exe main.go
GOOS=darwin  GOARCH=amd64  go build -o $GOPATH/bin/joss-darwin-x64  main.go
GOOS=linux   GOARCH=amd64  go build -o $GOPATH/bin/joss_linux_x64   main.go
GOOS=linux   GOARCH=386    go build -o $GOPATH/bin/joss_linux_x86   main.go
GOOS=linux   GOARCH=arm    go build -o $GOPATH/bin/joss-linux-arm main.go