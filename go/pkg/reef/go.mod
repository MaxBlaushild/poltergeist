module github.com/MaxBlaushild/poltergeist/pkg/reef

go 1.24.0

toolchain go1.24.10

replace (
	github.com/MaxBlaushild/poltergeist/pkg/aws => ../aws
	github.com/MaxBlaushild/poltergeist/pkg/email => ../email
)

require (
	github.com/MaxBlaushild/poltergeist/pkg/aws v0.0.0-00010101000000-000000000000
	github.com/MaxBlaushild/poltergeist/pkg/email v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
)

require (
	github.com/aws/aws-sdk-go v1.55.5 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)
