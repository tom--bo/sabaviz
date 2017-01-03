# sabaviz

sabaviz creates dot files(.dot) for graphviz, which visualize servers connections by netstat output.

## Description

## Usage
```
$ sabaviz -max 20 --exclude-processes ssh,ldap --exclude-ports 22 --host-check internal.domain.name target.host.name > graph.dot
```

Then you will get graph.dot for Graphviz.
To get image, use dot command.

For example
```
dot -Tpng graph.dot -o graph.png
```

![top-page](https://github.com/tom--bo/sabaviz/blob/image/sample.png)

## Install

To install, use `go get`:

```bash
$ go get -d github.com/tom--bo/sabaviz
```

## Contribution

1. Fork ([https://github.com/tom--bo/sabaviz/fork](https://github.com/tom--bo/sabaviz/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[tom--bo](https://github.com/tom--bo)
