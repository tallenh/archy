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
release version="":
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "{{version}}" ]; then
        latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "none")
        echo "Latest tag: $latest"
        exit 0
    fi
    if [ -n "$(git status --porcelain)" ]; then
        echo "error: working directory is not clean" >&2
        exit 1
    fi
    git push origin main
    git tag {{version}}
    git push origin {{version}}

# Promote a beta tag on HEAD to a stable release
promote:
    #!/usr/bin/env bash
    set -euo pipefail
    beta=$(git describe --tags --exact-match HEAD 2>/dev/null || true)
    if [ -z "$beta" ]; then
        echo "error: HEAD has no tag" >&2
        exit 1
    fi
    if [[ "$beta" != *-beta* ]]; then
        echo "error: tag '$beta' is not a beta tag" >&2
        exit 1
    fi
    stable=${beta%%-beta*}
    echo "Promoting $beta → $stable"
    gh release delete "$beta" --yes
    git tag -d "$beta"
    git push origin ":refs/tags/$beta"
    git tag "$stable"
    git push origin "$stable"

# Test release locally (no publish)
release-dry-run:
    goreleaser release --snapshot --clean
