# golocate

`golocate` is a utility for locating interface implementations, functions
and comments in go source code using the parsed representation of the code
rather than simple text search.

Locate all instances of io.Writer ./...
```sh
go run . --interfaces io.Writer ./...
```

Locate all exported functions in ./...
```sh
go run . --functions='.*' ./...
```

Locate all comments in ./...
```sh
go run . --comments='.*' ./...
```

The output of `golocate` is limited right now but is easily extended as
uses cases arise. Currently locating interface implementations is the
most useful.

