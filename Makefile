
export GOOGLE_APPLICATION_CREDENTIALS = gcp.json

GCP_PROJECT := $(shell python3 -c "import json; f = open('gcp.json'); j = json.load(f); print(j['project_id'])")
id_OPTIONS = --project $(GCP_PROJECT)

APP_ID = govanity

runlocal:
	go run server.go --debug

mongo:
	mongo `gcloud --project $(GCP_PROJECT) secrets versions access latest --secret=mongourl`

mongosetup:
	mongo `gcloud --project $(GCP_PROJECT) secrets versions access latest --secret=mongourl` mongosetup.js

rundocker: 
	docker build -t $(APP_ID) .
	docker run -p 8080:8080 -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/keys/FILE_NAME.json -v ${GOOGLE_APPLICATION_CREDENTIALS}:/tmp/keys/FILE_NAME.json:ro $(APP_ID)

