APPS     := init registry

all:	$(APPS)

bootstrap:
	hack/make-rules/tools.sh install

$(APPS):
	go build -o bin/pushup main.go

ent:
	go generate ./internal/adapters/persistence/ent/

openapi: openapi_http

openapi_http:
	@./scripts/openapi-http.sh openapi internal/adapters/driving/http server

openapi_js:
	@./scripts/openapi-js.sh openapi

clean:
	rm -rf bin/*

dep: 
	go get -v -d ./...
