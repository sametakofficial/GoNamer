
scan-project:
	@go run scripts/scan_project.go . -o project_knowledge.md

build:
	@go build -o build/gonamer ./cmd/