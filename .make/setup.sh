#!/usr/bin/env sh

set -ae

rm -f \
    .make/.gitignore \
    .make/setup.sh \
    .make/Makefile.common \
    .make/Makefile.revive \
    .make/revive.toml
mkdir -p .make
curl -sSL https://github.com/akaspin/make-go/tarball/master | tar --strip-components 1 -xz -C .make
