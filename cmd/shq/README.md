# `shq`

`printf %q` without the `printf`.  Quote your shell strings so you can pass
them to `eval` or otherwise run them somewhere else.

## Usage

```
shq s1 [s2 ... sN]
```

Will print its arguments quoted for `sh`.

## Example

```sh
CRAZYVAR_INPUT="this won't break"
echo "export CRAZYVAR=$(shq "$CRAZYVAR_INPUT")" > ~/.crazyvar

. ~/.crazyvar
echo "$CRAZYVAR"
```

prints

```text
this won't break
```
