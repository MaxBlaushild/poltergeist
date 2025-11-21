module github.com/MaxBlaushild/poltergeist/pkg/dropbox

go 1.24.0

require (
	github.com/MaxBlaushild/poltergeist/pkg/models v0.0.0-00010101000000-000000000000
	github.com/dropbox/dropbox-sdk-go-unofficial/v6 v6.0.5
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
)

replace (
	github.com/MaxBlaushild/poltergeist/pkg/models => ../models
)

