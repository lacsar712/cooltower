.PHONY: build test benzhi-docker stats
build:
	go build -o bin/cooltower ./cmd/cooltower
test:
	go test ./... -count=1
benzhi-docker:
	sh build_benzhi_docker.sh
stats:
	@powershell -NoProfile -Command "$$go = (Get-ChildItem -Recurse -Include *.go | Get-Content | Measure-Object -Line).Lines; $$files = (Get-ChildItem -Recurse -Include *.go).Count; Write-Host \"Go lines: $$go  files: $$files\""
