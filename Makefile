
BUILDDIR          := ${CURDIR}/_build
GOPATH            := ${BUILDDIR}
RPMBUILDDIR       := ${BUILDDIR}/rpm
GOXBUILDDIR       := ${BUILDDIR}/gox
TARBUILDDIR       := ${BUILDDIR}/tar
ARCH              := $(shell uname -m|tr '[:upper:]' '[:lower:]')
OS                := $(shell uname -s|tr '[:upper:]' '[:lower:]')
FPM_VERSION       := $(shell gem list fpm|grep fpm|sed -E 's/fpm \((.*)\)/\1/g')
FPM_OPTIONS       ?=
ITERATION         ?= 1

GPM               := ${CURDIR}/.tools/gpm
GOCAMO_VER        := $(shell git describe --always --dirty --tags|sed 's/^v//')
RPM_VER           := $(shell git describe --always --tags --abbrev=0|sed 's/^v//')
VERSION_VAR       := main.ServerVersion
GOTEST_FLAGS      := -cpu=1,2
GOBUILD_DEPFLAGS  := -tags netgo
GOBUILD_LDFLAGS   ?=
GOBUILD_FLAGS     := ${GOBUILD_DEPFLAGS} -ldflags "${GOBUILD_LDFLAGS} -X ${VERSION_VAR} ${GOCAMO_VER}"
GO                := env GOPATH="${GOPATH}" go
XCOMPILE_ARCHES   := darwin.amd64 freebsd.amd64 linux.amd64

define GO_CAMO_RPM_DESCRIPTION
Camo is a special type of image proxy that proxies non-secure images over
SSL/TLS. This prevents mixed content warnings on secure pages.
It works in conjunction with back-end code to rewrite image URLs and sign them
with an HMAC.
endef
export GO_CAMO_RPM_DESCRIPTION

define HELP_OUTPUT
Available targets:
  help                this help
  clean               clean up
  all                 build binaries and man pages
  test                run tests
  cover               run tests with cover output
  build               build all binaries
  man                 build all man pages
  tar                 build release tarball
  rpm                 build rpm
  cross-tar           cross compile and build release tarballs
endef
export HELP_OUTPUT

.PHONY: help clean build build-setup test cover man man-copy rpm all tar cross-tar

help:
	@echo "$$HELP_OUTPUT"

clean:
	-rm -rf "${BUILDDIR}"

build-setup:
	@mkdir -p "${GOPATH}/src"
	@echo "Restoring deps..."
	@env GOPATH="${GOPATH}" ${GPM} install
	@mkdir -p "${GOPATH}/src/github.com/cactus"
	@test -d "${GOPATH}/src/github.com/cactus/go-camo" || ln -s "${CURDIR}" "${GOPATH}/src/github.com/cactus/go-camo"

build: build-setup
	@echo "Building go-camo..."
	@${GO} install ${GOBUILD_FLAGS} github.com/cactus/go-camo
	@echo "Building url-tool..."
	@${GO} install ${GOBUILD_FLAGS} github.com/cactus/go-camo/url-tool
	@echo "Building simple-server..."
	@${GO} install ${GOBUILD_FLAGS} github.com/cactus/go-camo/simple-server

test: build-setup
	@echo "Running tests..."
	@${GO} test ${GOTEST_FLAGS} ./camo/... ./router/...

cover: build-setup
	@echo "Running tests with coverage..."
	@${GO} test -cover ${GOTEST_FLAGS} ./camo/... ./router/...

${BUILDDIR}/man/%: man/%.mdoc
	@mkdir -p "${BUILDDIR}/man"
	@cat $< | sed "s#.Os GO-CAMO VERSION#.Os GO-CAMO ${GOCAMO_VER}#" > $@

man: $(patsubst man/%.mdoc,${BUILDDIR}/man/%,$(wildcard man/*.1.mdoc))

tar: all
	@echo "Building tar..."
	@mkdir -p ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin
	@mkdir -p ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/man
	@cp ${BUILDDIR}/bin/go-camo ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/
	@cp ${BUILDDIR}/bin/simple-server ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/
	@cp ${BUILDDIR}/bin/url-tool ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/bin/
	@cp ${BUILDDIR}/man/* ${TARBUILDDIR}/go-camo-${GOCAMO_VER}/man/
	@tar -C ${TARBUILDDIR} -czf go-camo-${GOCAMO_VER}.${OS}.${ARCH}.tar.gz \
		go-camo-${GOCAMO_VER}
	@mv go-camo-${GOCAMO_VER}.${OS}.${ARCH}.tar.gz ${BUILDDIR}/

rpm: all
	@echo "Building rpm..."
	@mkdir -p ${RPMBUILDDIR}/usr/local/bin
	@mkdir -p ${RPMBUILDDIR}/usr/local/share/man/man1
	@cp ${BUILDDIR}/bin/{go-camo,simple-server,url-tool} \
		${RPMBUILDDIR}/usr/local/bin
	@cp ${BUILDDIR}/man/* ${RPMBUILDDIR}/usr/local/share/man/man1/
	@fpm -s dir -t rpm -n go-camo \
		--url "https://github.com/cactus/go-camo" \
		-v "${RPM_VER}" \
		--iteration "${ITERATION}" \
		--license MIT \
		--description "$$GO_CAMO_RPM_DESCRIPTION" \
		-C "${RPMBUILDDIR}" \
		${FPM_OPTIONS} \
		usr/local/bin usr/local/share/man/man1
	@mv *.rpm ${BUILDDIR}/

cross-tar: build-setup man
	@[ -f "${BUILDDIR}/bin/gox" ] || ${GO} get github.com/mitchellh/gox
	@${BUILDDIR}/bin/gox ${GOBUILD_FLAGS} \
		-osarch="freebsd/amd64 darwin/amd64 linux/amd64" \
		-output="${GOXBUILDDIR}/{{.OS}}.{{.Arch}}/{{.Dir}}" \
		github.com/cactus/go-camo \
		github.com/cactus/go-camo/url-tool \
		github.com/cactus/go-camo/simple-server
	@(cd ${GOXBUILDDIR}; for x in ${XCOMPILE_ARCHES}; do \
		echo "Making tar for go-camo.$${x}"; \
		mkdir -p ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/; \
		mkdir -p ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/man/; \
		cp $${x}/go-camo ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/; \
		cp $${x}/simple-server ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/; \
		cp $${x}/url-tool ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/bin/; \
		cp ${BUILDDIR}/man/* ${TARBUILDDIR}/$${x}/go-camo-${GOCAMO_VER}/man/; \
		tar -C ${TARBUILDDIR}/$${x} -czf go-camo-${GOCAMO_VER}.$${x}.tar.gz \
		   go-camo-${GOCAMO_VER}; \
		mv go-camo-${GOCAMO_VER}.$${x}.tar.gz ${BUILDDIR}/; \
	 done)

all: build man
