export GO = $(shell which go)
export GOOGLE_APPLICATION_CREDENTIALS = gcp.json

GCP_PROJECT := $(shell which python3 >/dev/null 2>&1 && test -f ${GOOGLE_APPLICATION_CREDENTIALS} >/dev/null 2>&1 && python3 -c "import json; f = open('${GOOGLE_APPLICATION_CREDENTIALS}'); j = json.load(f); print(j['project_id'])")
id_OPTIONS = --project $(GCP_PROJECT)

APP_ID = govanity

ifeq ($(DOCS_DIST),)
DOCS_DIST := $(shell pwd)/docs_dist.tgz
endif

goformat:
	go install golang.org/x/tools/cmd/goimports
	gofmt -s -w .
	goimports -w -local=go.viam.com/govanity .

lint: goformata
	go install github.com/edaniels/golinters/cmd/combined
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go vet -vettool=`go env GOPATH`/bin/combined
	go list -f '{{.Dir}}' ./... | grep -v gen | xargs go run github.com/golangci/golangci-lint/cmd/golangci-lint run -v

docs_bootstrap:
	./etc/docs_bootstrap.sh

docs_static: docs_bootstrap
	./etc/docs_static.sh ${DOCS_DIST}

runlocal:
	go run cmd/server/main.go --debug

runlocaldummy:
	GO_MODULES=vanity.domain.com/private-repo=github.com/org/private-repo,github.com/org/no-vanity-private-repro go run cmd/server/main.go --debug

mongo:
	mongo `gcloud --project $(GCP_PROJECT) secrets versions access latest --secret=mongourl`

mongo_setup:
	mongo `gcloud --project $(GCP_PROJECT) secrets versions access latest --secret=mongourl` mongosetup.js

build_docker:
	docker build --build-arg GO_MODULES=${GO_MODULES} -t $(APP_ID) .

run_docker: build_docker 
	docker run -p 8080:8080 -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/keys/FILE_NAME.json -v ${GOOGLE_APPLICATION_CREDENTIALS}:/tmp/keys/FILE_NAME.json:ro $(APP_ID)
