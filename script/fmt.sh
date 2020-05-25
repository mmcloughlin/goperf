#!/bin/bash -ex

repo="github.com/mmcloughlin/goperf"

# Files to apply formatting to.
files=$(find . -name '*.go' -not -path '*/vendor/*')

# Remove blank lines in import blocks. This will force formatting to group
# imports correctly.
sed -i.fmtbackup '/^import (/,/)/ { /^$/ d; }' ${files}
find . -name '*.fmtbackup' -delete

# goimports is goimports with stricter formatting.
gofumports -w -local ${repo} ${files}
gofumpt -w -s ${files}
