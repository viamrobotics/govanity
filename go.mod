module app

go 1.15

//replace github.com/erh/egoutil => /Users/erh/work/egoutil

require (
	contrib.go.opencensus.io/exporter/stackdriver v0.13.4
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/edaniels/golog v0.0.0-20210104162753-3254576d0129
	github.com/erh/egoutil v0.0.10
	go.mongodb.org/mongo-driver v1.4.5
	go.opencensus.io v0.22.5
)
