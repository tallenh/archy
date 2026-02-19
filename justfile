default: build

build:
    go build -o archy ./cmd/archy

test:
    go test ./...

vet:
    go vet ./...

clean:
    rm -f archy

check: vet test build

# Tag and push a release (e.g. just release v0.2.0)
release version:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -n "$(git status --porcelain)" ]; then
        echo "error: working directory is not clean" >&2
        exit 1
    fi
    git push origin main
    git tag {{version}}
    git push origin {{version}}

# Test release locally (no publish)
release-dry-run:
    goreleaser release --snapshot --clean
