#!/bin/bash -ex

# Install golangci-lint
golangci_lint_version='v1.23.7'
curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $GOPATH/bin ${golangci_lint_version}

# Install Go tools.
GO111MODULE=off go get -u github.com/myitcv/gobin
gobin mvdan.cc/gofumpt@v0.0.0-20200412215918-a91da47f375c
gobin mvdan.cc/gofumpt/gofumports@v0.0.0-20200412215918-a91da47f375c
gobin github.com/alvaroloes/enumer@v1.1.2
gobin github.com/kyleconroy/sqlc/cmd/sqlc@v1.3.0
gobin github.com/GoogleCloudPlatform/cloudsql-proxy/cmd/cloud_sql_proxy@v0.0.0-20200325185443-f6b3391c52cf
