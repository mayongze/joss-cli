GIT_SHA=`git rev-parse --short HEAD || echo`
VERSION=`git describe --tags`

build:
	@echo "Building joss-cli..."
	@mkdir -p bin
	@go build -ldflags "-X github.com/mayongze/joss-cli/command.GitSHA=${GIT_SHA} -X github.com/mayongze/joss-cli/command.Version=${VERSION}" -o bin/joss-cli .

install:
	@echo "Installing joss-cli..."
	@install -c bin/joss-cli /usr/local/bin/joss-cli

clean:
	@rm -f bin/*
