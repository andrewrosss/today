# Today

```bash
go install github.com/andrewrosss/today/src/go/today@v0.2.0
```

## Releasing a new version

In the directory `src/go/today`, from a clean working directory/commit run:

```bsah
go mod tidy
```

Commit any changes that `go mod tidy` made.

Then bump the version with `bump2version`:

```bash
# we use bump2version so that the version number is defined statically
# in the source code and managed/incremented through a single tool
# in particular, this ensures consistency between:
# - the version number in the source code
# - the version number in git tag
# - the version number in the README
# - the version number/format of the new version's commit message
pipx run bump2version --new-versio='x.y.z' fakepart
```

Then push the tags:

```bash
git push --tags
```

Then make the module available by running the go list command to prompt Go to update its index of modules with information about the module you’re publishing.

```bash
# see: https://stackoverflow.com/a/7979255/2889677 for more information
# on the incantation below to get the latest git tag, i.e.:
# `git describe --tags $(git rev-list --tags --max-count=1)`
GOPROXY=proxy.golang.org go list -m "$(git describe --tags $(git rev-list --tags --max-count=1) | __MODULE=$(grep -E '^module' go.mod | cut -d' ' -f2) sed "s+.*today/+$__MODULE@+")"
```

:trophy: Done!
