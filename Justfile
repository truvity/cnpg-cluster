# Development commands for cnpg-cluster (chart-only repository)

# Disable go.work (parent workspace interferes with standalone module builds)
export GOWORK := "off"

# Format all Go files (gofmt + goimports via golangci-lint)
fmt:
    golangci-lint fmt ./...

# Run the chart render tests (chart_test.go drives `helm template`)
test:
    go test ./... -coverprofile=coverage.out

# Run linters
lint:
    golangci-lint run ./...

# Run Go vulnerability check
vuln:
    govulncheck ./...

# Run go mod tidy
tidy:
    go mod tidy

# Clean build artifacts
clean:
    rm -rf dist/ coverage.out

# Lint + render both charts. cnpg-cluster's deep posture assertions live
# in chart_test.go (recipe `test`); this recipe proves both charts lint
# and render from their defaults, and that cnpg-database renders with
# the minimum required values.
chart-lint:
    helm lint charts/cnpg-cluster --set clusterName=pg --set namespace=default --set profile=devel
    helm template pg charts/cnpg-cluster \
        --set clusterName=pg --set namespace=default --set profile=devel >/dev/null
    helm lint charts/cnpg-database \
        --set clusterName=pg --set namespace=default --set profile=devel --set databaseName=app
    helm template db charts/cnpg-database \
        --set clusterName=pg --set namespace=default --set profile=devel --set databaseName=app >/dev/null

check: test lint chart-lint vuln

# Build a snapshot release locally (no push, no tag)
snapshot:
    goreleaser release --snapshot --clean

# Package Helm charts locally
helm-package:
    helm package charts/cnpg-cluster --destination dist/
    helm package charts/cnpg-database --destination dist/
