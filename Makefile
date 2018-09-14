# workarounds

mod-vendor: vendor/modules.txt ## IDEA don't want to index

vendor/modules.txt: go.mod
	GO111MODULE=on go mod vendor
