#!/usr/bin/env sh

set -ae

GO111MODULES=on

cp go.mod go.mod.bak
go get -u
go mod tidy
mv go.mod go.mod.up
mv go.mod.bak go.mod
diff go.mod go.mod.up
