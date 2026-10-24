# set some just variables
j_dir := justfile_directory()
build_dir := clean(j_dir / "build")
tarout_dir := clean(build_dir / "tar")

# build specific vars
APP_NAME := env("APP_NAME", "go-camo")
APP_VER := env("APP_VER", "")
GITHASH := env("GITHASH", "")
BUILD_OPTIONS := env("BUILD_OPTIONS", "-trimpath")
BUILD_DEPFLAGS := env("BUILD_DEPFLAGS", '-tags netgo,production')
BUILD_LDFLAGS := env("BUILD_LDFLAGS", '-s -w')
SIGN_KEY := env("SIGN_KEY", home_dir() / ".minisign/go-camo.key")
CC_BUILD_TARGETS := env("CC_BUILD_TARGETS", "go-camo url-tool")
CC_BUILD_ARCHES := env("CC_BUILD_ARCHES", "darwin-arm64 freebsd-amd64 linux-amd64 linux-arm64 windows-amd64")
GOVER := `go version | awk '{print $3}' | tr -d '.'`
OS := `go env GOHOSTOS`
ARCH := `go env GOHOSTARCH`

# export some build specific env vars
export VERSION_VAR := "main.ServerVersion"

# clean
[group('build')]
clean: (_banner ">> cleaning build dir")
    rm -rf "{{ build_dir }}"

# build: manpages
[group('build')]
manpages: (_banner ">> building manual pages")
    [ -d "{{ build_dir }}/man" ] || mkdir -p "{{ build_dir }}/man"
    for file in man/*.scd; do \
        bname="${file##*/}"; \
        fname="${bname%.scd}"; \
        scdoc < $file > "{{ build_dir }}/man/${fname}"; \
    done

# build: binaries for current architecture
[group('build')]
build: (_banner ">> building binaries")
    #!/bin/sh
    set -eu
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    [ -d "{{ build_dir }}/bin" ] || mkdir -p "{{ build_dir }}/bin"
    echo "> Building as APP_VER=${APP_VER}"
    for x in {{ CC_BUILD_TARGETS }}; do
        echo " * ${x} ...";
        go build {{ BUILD_DEPFLAGS }} {{ BUILD_OPTIONS }} \
          -ldflags="{{ BUILD_LDFLAGS }} -X {{ VERSION_VAR }}=${APP_VER}" \
          -o "{{ build_dir }}/bin/${x}" ./cmd/${x};
    done

# build: binaries for current architecture; manpages; assemble tarball
[group('release')]
tar: manpages build (_banner ">> assembling tarball")
    #!/bin/sh
    set -eu
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    OUTDIR="{{ tarout_dir }}/{{ APP_NAME }}-${APP_VER}"
    mkdir -p "${OUTDIR}"
    cp -R {{ build_dir }}/bin ${OUTDIR}/
    cp -R {{ build_dir }}/man ${OUTDIR}/
    tar -C {{ tarout_dir }} \
      -czf {{ tarout_dir }}/{{ APP_NAME }}-${APP_VER}.{{ GOVER }}.{{ OS }}-{{ ARCH }}.tar.gz \
      {{ APP_NAME }}-${APP_VER}/
    rm -rf "${OUTDIR}"

# build: cross compiled binaries
[group('build')]
cross-build: (_banner ">> building cc targets")
    #!/bin/sh
    set -eu
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    for x in {{ CC_BUILD_TARGETS }}; do
        for y in {{ CC_BUILD_ARCHES }}; do
            printf -- "--> %15s: %s\n" "${y}" "${x}"
            GOOS="${y%%-*}"
            GOARCH="${y##*-}"
            EXT=""
            if echo "${y}" | grep -q 'windows-'; then EXT=".exe"; fi;
            env GOOS=${GOOS} GOARCH=${GOARCH} \
                go build {{ BUILD_DEPFLAGS }} {{ BUILD_OPTIONS }} \
                  -ldflags "{{ BUILD_LDFLAGS }} -X {{ VERSION_VAR }}=${APP_VER}" \
                  -o "{{ build_dir }}/bin/${x}.${GOOS}-${GOARCH}${EXT}" \
                  ./cmd/${x}
        done
    done

# build: cross compile binaries; manpages; assemble tarballs
[group('release')]
cross-tar: manpages cross-build (_banner ">> building cc tarballs")
    #!/bin/sh
    set -eu
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    mkdir -p "{{ tarout_dir }}"
    for x in {{ CC_BUILD_ARCHES }}; do
        printf -- "--> %15s\n" "${x}"
        EXT=""
        if echo "${x}" | grep -q 'windows-'; then EXT=".exe"; fi;
        XDIR="{{ GOVER }}.${x}"
        ODIR="{{ tarout_dir }}/${XDIR}/{{ APP_NAME }}-${APP_VER}"
        mkdir -p "${ODIR}/bin"; \
        mkdir -p "${ODIR}/man"; \
            for t in {{ CC_BUILD_TARGETS }}; do
                cp {{ build_dir }}/bin/${t}.${x}${EXT} ${ODIR}/bin/${t}${EXT}
            done
        cp -R {{ build_dir }}/man ${ODIR}/
        tar -C {{ tarout_dir }}/${XDIR} \
          -czf {{ tarout_dir }}/{{ APP_NAME }}-${APP_VER}.${XDIR}.tar.gz \
          {{ APP_NAME }}-${APP_VER}
        rm -rf "{{ tarout_dir }}/${XDIR}/"
    done

# build: container images
[group('build')]
build-oci-image *FLAGS: (_banner ">> building container images")
    #!/bin/sh
    set -eu
    GITHASH="{{ GITHASH }}"
    if [ "$GITHASH" = "" ]; then
        GITHASH="$(git rev-parse --short HEAD)"
    fi
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    docker build {{ FLAGS }} \
      --build-arg GITHASH=${GITHASH} \
      --build-arg APP_VER=${APP_VER} \
      --build-arg BUILDKIT_INLINE_CACHE=1 \
      -t go-camo:latest \
      -f docker/Dockerfile \
      .

# sign release tarballs
[group('release')]
release-sign: (_banner ">> signing release")
    #!/bin/sh
    set -eu
    APP_VER="{{ APP_VER }}"
    if [ "$APP_VER" = "" ]; then
        APP_VER="$(git describe --always --tags|sed 's/^v//')"
    fi
    (
        cd build/tar; shasum -a 256 go-camo-*.tar.gz > SHA256; \
        minisign -S -s {{ SIGN_KEY }} \
          -c "go-camo-${APP_VER} SHA256" \
          -t "go-camo-${APP_VER} SHA256" \
          -x SHA256.sig -m SHA256; \
    )

# build: cross compile binaries; assemble tarballs; sign tarballs
[group('release')]
release: cross-tar release-sign