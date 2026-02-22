APPS     := init registry

all:	$(APPS)

bootstrap:
	hack/make-rules/tools.sh install

$(APPS):
	go build -o bin/server main.go

ent:
	go generate ./internal/adapters/driven/repository/ent/

mocks:
	mockery

.PHONY: openapi
openapi:
	npx @redocly/cli bundle openapi/openapi.yml -o openapi/bundled.yml
	oapi-codegen --config openapi/codegen_types.yml openapi/bundled.yml
	oapi-codegen --config openapi/codegen_server.yml openapi/bundled.yml
	rm openapi/bundled.yml

clean:
	rm -rf bin/*

dep: 
	go get -v -d ./...
