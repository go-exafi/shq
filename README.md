# `go-shq`

Ergonomic shell quoting with golang fmt.

## Description

Adds a type annotation, `shq.Arg`, which can be added to a string or []byte
to cause it to be shell quoted when it is used via `fmt.Stringer`.  This has
the effect that you can safely do something like the following:

```go
// add the annotation immediately upon entry into memory so we can't
// accidentally use them in their raw form.
input1 := shq.Arg("this isn't a safe string")
input2 := shq.Arg("what\nwill\nhappen???!")
exec.Command("sh", "-c", fmt.Sprintf("echo %s %s", input1, input2))
```

See the docs for more thorough examples.

## License

See [LICENSE](./LICENSE) file.
