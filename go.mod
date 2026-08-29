module github.com/51Degrees/ip-intelligence-go/v4

go 1.22

require (
	github.com/51Degrees/common-go/v4 v4.5.0
	github.com/51Degrees/pipeline-go v0.0.0
	github.com/goccy/go-yaml v1.19.2
	golang.org/x/text v0.21.0
)

// pipeline-go is not yet published. For local development it is resolved from a
// sibling checkout via a go.work workspace; add a replace directive (or publish
// pipeline-go and pin the require above) to build this module on its own.
