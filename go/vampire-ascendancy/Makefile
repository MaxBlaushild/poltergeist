dev:
	go run ./cmd/server/main.go --config-name local

seed:
	go run ./cmd/seed --config-name local --file seed/characters.json

# Regenerate seed/characters.json from the master packet PDF.
extract:
	python3 seed/extract.py "$(PDF)" seed/characters.json
