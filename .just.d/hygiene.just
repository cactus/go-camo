# run 'go fix'
[group('hygiene')]
fix: (_banner ">> running go fix")
    go fix -rangeint=false ./...

# update go.mod dependencies
[group('hygiene')]
update-go-deps: (_banner ">> updating go.mod dependencies")
    go get -u all
    go mod tidy
    go mod vendor

# update mise dependencies
[group('hygiene')]
update-mise-deps: (_banner ">> updating mise dependencies")
    mise up -b
    mise lock

# update dependencies
[group('hygiene')]
update-deps:
    just update-mise-deps
    just update-go-deps
