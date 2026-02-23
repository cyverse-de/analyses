default: build

build: docs analyses

analyses:
    go build -o bin/analyses cmd/analyses/*.go

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
