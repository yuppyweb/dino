install: 
    go mod tidy -x
    go install tool

lint:
    golangci-lint fmt
    golangci-lint --color=always run

test:
    go test ./... --count=1 -v -coverprofile cover.out && \
    go tool cover -html cover.out -o cover.html && \
    go tool cover -func cover.out && \
    rm cover.out

vuln:
    govulncheck -show=color -test ./...
