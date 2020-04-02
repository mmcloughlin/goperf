#!/bin/bash -ex

# Install golangci-lint
golangci_lint_version='v1.23.7'
curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $GOPATH/bin ${golangci_lint_version}

# Install Go tools.
GO111MODULE=off go get -u \
    mvdan.cc/gofumpt/gofumports \
    github.com/kyleconroy/sqlc/cmd/sqlc \
    github.com/GoogleCloudPlatform/cloudsql-proxy/cmd/cloud_sql_proxy
