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
