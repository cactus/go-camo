
BUILDDIR          := ${CURDIR}/build
GOPATH            := ${BUILDDIR}
GODEP             := ${GOPATH}/bin/godep
RPMBUILDDIR       := ${BUILDDIR}/rpm
ARCH              := $(shell uname -m)
FPM_VERSION       := $(shell gem list fpm|grep fpm|sed -E 's/fpm \((.*)\)/\1/g')
FPM_OPTIONS       ?=
ITERATION         ?= 1

GOCAMO_VER        := $(shell git describe --always --dirty --tags|sed 's/^v//')
RPM_VER           := $(shell git describe --always --tags --abbrev=0|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOTEST_FLAGS      :=
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := $(GOBUILD_DEPFLAGS) -ldflags "$(GOBUILD_LDFLAGS) -X $(VERSION_VAR) $(GOCAMO_VER)"

.PHONY: help clean build test cover man man-copy rpm all

help:
	@echo "Available targets:"
	@echo "  help                this help"
	@echo "  clean               clean up"
	@echo "  all                 build binaries and man pages"
	@echo "  build               build all"
	@echo "  build-go-camo       build go-camo"
	@echo "  build-url-tool      build url tool"
	@echo "  build-simple-server build simple server"
	@echo "  test                run tests"
	@echo "  cover               run tests with cover output"
	@echo "  man                 build all man pages"
	@echo "  man-go-camo         build go-camo man pages"
	@echo "  man-url-tool        build url-tool man pages"
	@echo "  man-simple-server   build simple-server man pages"
	@echo "  rpm                 build rpm"

clean:
	-rm -rf "${BUILDDIR}"

${GODEP}:
	@mkdir -p "${GOPATH}/src"
	@echo "Building godep..."
	@env GOPATH="${GOPATH}" go get ${GOBUILD_DEPFLAGS} github.com/kr/godep

build-setup: ${GODEP}
	@echo "Restoring deps with godep..."
	@env GOPATH="${GOPATH}" ${GODEP} restore
	@mkdir -p "${GOPATH}/src/github.com/cactus"
	@test -d "${GOPATH}/src/github.com/cactus/go-camo" || ln -s "${CURDIR}" "${GOPATH}/src/github.com/cactus/go-camo"

build-go-camo: build-setup
	@echo "Building go-camo..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} github.com/cactus/go-camo

build-url-tool: build-setup
	@echo "Building url-tool..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} github.com/cactus/go-camo/url-tool

build-simple-server: build-setup
	@echo "Building simple-server..."
	@env GOPATH="${GOPATH}" go install ${GOBUILD_FLAGS} github.com/cactus/go-camo/simple-server

test: build-setup
	@echo "Running tests..."
	@env GOPATH="${GOPATH}" go test ${GOTEST_FLAGS} ./camo/...

cover: build-setup
	@echo "Running tests with coverage..."
	@env GOPATH="${GOPATH}" go test -cover ${GOTEST_FLAGS} ./camo/...

${BUILDDIR}/man/man1/%.1: man/%.mdoc
	@mkdir -p "${BUILDDIR}/man/man1"
	@cp $< $@

man-camo: ${BUILDDIR}/man/man1/go-camo.1
man-url-tool: ${BUILDDIR}/man/man1/url-tool.1
man-simple-server: ${BUILDDIR}/man/man1/simple-server.1

rpm: all
	@echo "Building rpm..."
	@mkdir -p ${RPMBUILDDIR}/usr/local/bin
	@mkdir -p ${RPMBUILDDIR}/usr/local/share/man/man1
	@cp ${BUILDDIR}/bin/{go-camo,simple-server,url-tool} \
		${RPMBUILDDIR}/usr/local/bin
	@cp ${BUILDDIR}/man/man1/* ${RPMBUILDDIR}/usr/local/share/man/man1
	@fpm -s dir -t rpm -n go-camo \
		--url "https://github.com/cactus/go-camo" \
		-v "${RPM_VER}" \
		--iteration "${ITERATION}" \
		--license MIT \
		--description "Camo is a special type of image proxy that proxies non-secure images over SSL/TLS. This prevents mixed content warnings on secure pages.\nIt works in conjunction with back-end code to rewrite image URLs and sign them with an HMAC." \
		-C "${RPMBUILDDIR}" \
		${FPM_OPTIONS} \
		usr/local/bin usr/local/share/man/man1
	@mv *.rpm ${BUILDDIR}/

build: build-go-camo build-url-tool build-simple-server
man: man-camo man-url-tool man-simple-server
all: build man
