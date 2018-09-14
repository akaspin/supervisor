##
## workaround to use vendor path with go 1.11 because
## Intellij IDEA Go still works horrible with modules
##

mod-vendor: vendor/modules.txt

vendor/modules.txt: go.mod
	GO111MODULE=on go mod vendor
