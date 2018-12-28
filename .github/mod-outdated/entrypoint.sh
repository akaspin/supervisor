#!/usr/bin/env sh

set -ae

GO111MODULE=on

rm -f go.sum
cp go.mod go.mod.bak
go get -u
go mod tidy
mv go.mod go.mod.up
mv go.mod.bak go.mod
diff go.mod go.mod.up

