# go-make

The goal is simplify and streamline development of Golang projects.

## Installation and update 

```console
$ curl -sSL https://github.com/akaspin/make-go/raw/master/setup.sh | sh -
```

This oneliner will install files to `.make` directory. To use features include
installed Makefiles in your Make.

```makefile
include .make/Makefile.common .make/Makefile.shadow
```

Use `make-go-update` target to update to the latest version.

