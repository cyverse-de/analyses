default: build

# Pass ldflags to strip debug info in production builds:
#   just ldflags='-w -s' build
ldflags := ''

build: docs analyses

analyses:
    go build {{ if ldflags != "" { "-ldflags=" + ldflags } else { "" } }} -o bin/analyses cmd/analyses/*.go

test:
    go test ./...

fmt-docs:
    swag fmt -g app.go -d cmd/analyses/,httphandlers/,common/

docs: fmt-docs
    swag init --parseDependency -g app.go -d cmd/analyses/,httphandlers/,common/

clean:
    #!/usr/bin/env bash
    go clean
    if [ -f bin/analyses ]; then
        rm bin/analyses
    fi
