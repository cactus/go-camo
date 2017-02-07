
BUILDDIR          := ${CURDIR}
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell go env GOHOSTARCH)
OS                := $(shell go env GOHOSTOS)
GOVER             := $(shell go version | awk '{print $$3}' | tr -d '.')
APP_NAME          := go-camo
APP_VER           := $(shell git describe --always --dirty --tags|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR}=${APP_VER}"
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

clean-vendor:
	@rm -rf "${BUILDDIR}/vendor/src"

build-setup:
	@go get github.com/constabulary/gb/...
	@${GB} vendor restore

build:
	@echo "Building..."
	@${GB} build ${GOBUILD_FLAGS} ...

test:
	@echo "Running tests..."
	@${GB} test ${GOTEST_FLAGS} ...

generate:
	@echo "Running generate..."
	@${GB} generate

cover:
	@echo "Running tests with coverage..."
	@${GB} test -cover ${GOTEST_FLAGS} ...

${BUILDDIR}/man/%: man/%.mdoc
	@cat $< | sed -E "s#.Os (.*) VERSION#.Os \1 ${APP_VER}#" > $@

man: $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

tar: all
	@echo "Building tar..."
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin
	@mkdir -p ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man
	@cp ${BUILDDIR}/bin/${APP_NAME}-netgo ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin/${APP_NAME}
	@cp ${BUILDDIR}/bin/url-tool-netgo ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/bin/url-tool
	@cp ${BUILDDIR}/man/*.[1-9] ${TARBUILDDIR}/${APP_NAME}-${APP_VER}/man/
	@tar -C ${TARBUILDDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.${GOVER}.${OS}-${ARCH}.tar.gz ${APP_NAME}-${APP_VER}

cross-tar: man
	@echo "Making tar for ${APP_NAME}:darwin.amd64"
	@env GOOS=darwin  GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...
	@env GOOS=freebsd GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...
	@env GOOS=linux   GOARCH=amd64 ${GB} build ${GOBUILD_FLAGS} ...

	@(for x in darwin-amd64 freebsd-amd64 linux-amd64; do \
		echo "Making tar for ${APP_NAME}.$${x}"; \
		XDIR="${GOVER}.$${x}"; \
		mkdir -p ${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}/bin/; \
		mkdir -p ${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}/man/; \
		cp bin/${APP_NAME}-$${x}-netgo ${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}/bin/${APP_NAME}; \
		cp bin/url-tool-$${x}-netgo ${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}/bin/url-tool; \
		cp ${BUILDDIR}/man/*.[1-9] ${TARBUILDDIR}/$${XDIR}/${APP_NAME}-${APP_VER}/man/; \
		tar -C ${TARBUILDDIR}/$${XDIR} -czf ${TARBUILDDIR}/${APP_NAME}-${APP_VER}.$${XDIR}.tar.gz ${APP_NAME}-${APP_VER}; \
	done)

release-sign:
	@echo "signing release tarballs"
	@(cd tar; shasum -a 256 go-camo-*.tar.gz > SHA256; \
	  signify -S -s $${SECKEY} -m SHA256; \
	  sed -i.bak -E 's#^(.*:).*#\1 go-camo-${APP_VER} SHA256#' SHA256.sig; \
	  rm -f SHA256.sig.bak; \
	 )

all: build man
