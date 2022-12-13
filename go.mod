module github.com/vladopajic/go-actor

go 1.19

require (
	github.com/gammazero/deque v0.2.0
	github.com/stretchr/testify v1.8.0
	go.uber.org/goleak v1.1.12
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// versioning changed to 0.x.x
retract [v1.0.0, v1.5.0]
