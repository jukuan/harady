# shellcheck shell=bash
# Sourced by every bin/*.sh script. Ensures $GOBIN (~/go/bin) is on PATH
# within the current process, regardless of what the parent shell has.
# Also quiets any shadowing older copies (e.g. a Homebrew golangci-lint).
_harady_gobin="$(go env GOPATH)/bin"
case ":$PATH:" in
    *":$_harady_gobin:"*) ;;
    *) PATH="$_harady_gobin:$PATH"; export PATH ;;
esac
unset _harady_gobin
