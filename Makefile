default: build

GO := $(shell which go)
BIN := app
LDFLAGS := -ldflags="-s -w"

build:
	$(GO) build  -o bin/${BIN} main.go
	@mkdir -p bin

run:
	@chmod a+x bin/${BIN}
	bin/${BIN}


clean:
	@m -rf bin
