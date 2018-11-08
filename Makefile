GIT_SHA=`git rev-parse --short HEAD || echo`

build:
	@echo "Building joss-cli..."
	@mkdir -p bin
	@go build -ldflags "-X command.GitSHA=${GIT_SHA}" -o bin/joss-cli .

install:
	@echo "Installing joss-cli..."
	@install -c bin/joss-cli /usr/local/bin/joss-cli

clean:
	@rm -f bin/*

dep:
	@dep ensure