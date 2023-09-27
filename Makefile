
APP_NAME=chat

VERSION=$(shell git describe --abbrev=0 --tags 2>/dev/null || echo "v0.0.0")
COMMIT=$(shell git rev-parse --short HEAD || echo "HEAD")
MODULE=$(shell grep ^module go.mod | awk '{print $$2;}')
LD_FLAGS="-w -X $(MODULE)/goapp.Version=$(VERSION) -X $(MODULE)/goapp.Commit=$(COMMIT)"
MAIN="goapp/service/main/main.go"

run: binary
	@DEV=1 ADDRESS=0.0.0.0:8080 ./temp/$(APP_NAME)

deploy: binary
	scp temp/chat optiplex:/tmp/chat
	ssh optiplex sudo /tmp/chat -action deploy

binary: wasm
	@mkdir -p temp
	@CGO_ENABLED=0 go build -o temp/$(APP_NAME) -ldflags $(LD_FLAGS) $(MAIN)

wasm:
	@rm -f internal/goapp/web/app.wasm
	@GOARCH=wasm GOOS=js go build -o goapp/web/app.wasm -ldflags $(LD_FLAGS) $(MAIN)

clean:
	@rm -rf temp
	@rm -f goapp/web/app.wasm

