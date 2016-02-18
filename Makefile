
BUILDDIR          := ${CURDIR}
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell go env GOHOSTARCH)
OS                := $(shell go env GOHOSTOS)
GOCAMO_VER        := $(shell git describe --always --dirty --tags|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${GOCAMO_VER}"
GB                := gb

define HELP_OUTPUT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  test                run tests
  cover               run tests with cover output
  build-setup         fetch dependencies
  build               build all binaries
  man                 build all man pages
  tar                 build release tarball
  cross-tar           cross compile and build release tarballs
endef
export HELP_OUTPUT

.PHONY: help clean build test cover man man-copy all tar cross-tar

help:
	@echo "$$HELP_OUTPUT"

clean:
	@rm -rf "${BUILDDIR}/bin"
	@rm -rf "${BUILDDIR}/pkg"
	@rm -rf "${BUILDDIR}/tar"
	@rm -rf "${BUILDDIR}/man/"*.[1-9]

build-setup:
	@go get github.com/constabulary/gb/...
	@${GB} vendor restore

build:
	@echo "Building go-camo..."
	@${GB} build ${GOBUILD_FLAGS} ...

test:
	@echo "Running tests..."
	@${GB} test ${GOTEST_FLAGS} ...

cover:
	@echo "Running tests with coverage..."
	@${GB} test -cover ${GOTEST_FLAGS} ...

${BUILDDIR}/man/%: man/%.mdoc
	@cat $< | sed "s#.Os GO-CAMO VERSION#.Os GO-CAMO ${GOCAMO_VER}#" > $@

man: $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

tar: all
	@echo "Building tar..."
	@mkdir -p ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin
	@mkdir -p ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/man
	@cp ${BUILDDIR}/bin/go-camo-netgo ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/go-camo
	@cp ${BUILDDIR}/bin/simple-server-netgo ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/simple-server
	@cp ${BUILDDIR}/bin/url-tool-netgo ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/url-tool
	@cp ${BUILDDIR}/man/*.[1-9] ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/man/
	@tar -C ${TARBUILDDIR} -czf ${TARBUILDDIR}/go-camo-${GOCAMO_VER}.${OS}.${ARCH}.tar.gz go-camo-${GOCAMO_VER}

cross-tar: man
	@echo "Making tar for go-camo:darwin.amd64"
	@env GOOS=darwin  GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...
	@env GOOS=freebsd GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...
	@env GOOS=linux   GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...

	@(for x in darwin-amd64 freebsd-amd64 linux-amd64; do \
		echo "Making tar for go-camo.$${x}"; \
		mkdir -p ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/; \
		mkdir -p ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/man/; \
		cp bin/go-camo-$${x}-netgo ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/go-camo; \
		cp bin/simple-server-$${x}-netgo ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/simple-server; \
		cp bin/url-tool-$${x}-netgo ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/url-tool; \
		cp ${BUILDDIR}/man/* ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/man/; \
		tar -C ${TARBUILDDIR}/$${x} -czf ${TARBUILDDIR}/go-camo-${GOCAMO_VER}.$${x}.tar.gz go-camo-${GOCAMO_VER}; \
	done)

all: build man
