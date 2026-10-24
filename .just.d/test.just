# test specific vars
GOTEST_FLAGS := env("GOTEST_FLAGS", "-cpu=1,2 -count=1 -vet=off")
GOTESTSUM_FLAGS := env("GOTESTSUM_FLAGS", "")

# run coverage analysis
[group('tests')]
cover:
    just _banner ">> running coverage analysis"
    go test -cover {{GOTEST_FLAGS}} ./...

# run benchmarks
[group('tests')]
bench +FLAGS='-vet=off':
    just _banner ">> running benchtests"
    go test -bench="." -run="^$" -test.benchmem=true {{FLAGS}} ./...

# run fuzz tests
[group('tests')]
fuzz TEST_NAME +FLAGS:
    just _banner ">> running benchtests"
    go test -fuzz={{TEST_NAME}} {{FLAGS}}

# run tests
[group('tests')]
test:
    just _banner ">> running tests"
    gotestsum {{GOTESTSUM_FLAGS}} -- {{GOTEST_FLAGS}} ./...
