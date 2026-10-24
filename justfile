# this setting will allow passing arguments through to tasks, see the docs here
# https://just.systems/man/en/settings.html#positional-arguments
set positional-arguments
# Disable echoing recipe lines before executing
set quiet
# Enable unstable features
set unstable
# Ignore recipe lines beginning with "#"
set ignore-comments
# causes the evaluator to skip evaluating unused variables
set lazy
# Values may be lists of strings instead of strings
set lists

# import the user's justfile if it exists
import? '~/.justfile'
import? '~/justfile'

# import split up files
import '.just.d/build.just'
import '.just.d/check.just'
import '.just.d/hygiene.just'
import '.just.d/test.just'

# some justfile global exports
export GOPRIVATE := "github.com/dropwhile/go-camo"
# unsure toolchains are not automatically downloaded/upgraded
export GOTOOLCHAIN := "local"
# disable cgo
export CGO_ENABLED := "0"

# show the list of available commands
[default]
help:
    just --list --justfile {{ justfile() }}

_banner *ARGS:
    printf '{{ YELLOW }}[%s] %-72s{{ NORMAL }}\n' "$(date +%H:%M:%S)" "{{ ARGS }}"
