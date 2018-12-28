#!/usr/bin/env sh

set -ae

GO111MODULE=on

rm -f go.sum
cp go.mod /tmp/go.mod.bak
go get -u
go mod tidy
mv go.mod /tmp/go.mod.up
mv /tmp/go.mod.bak go.mod
diff go.mod /tmp/go.mod.up

