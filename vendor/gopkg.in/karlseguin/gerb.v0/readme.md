# Gerb
Gerb is a erb-inspired templating engine for Go.

## Usage

```go
template, err := gerb.ParseString(true, "....")
if err != nil {
  panic(err)
}
data := map[string]interface{}{
  "name": ....
}
template.Render(os.Stdout, data)
```

There are three available methods for creating a template:

1. `Parse(cache bool, data [][]byte)`
2. `ParseString(cache bool, data []string)`
3. `ParseFile(cache bool, paths []string)`

Unless `cache` is set to `false`, an internal cache is used to avoid having
to parse the same content (based on the content's hash). The cache will
automatically evict old items.

Once you have a template, you use the `Render` method to output the template
to the specified `io.Writer` using the specified data. Data must be a
`map[string]interface{}` or nil in the rare case where you have no template data.

It's safe to call `Render` from multiple threads.

## Output Tags
Gerb supports two types of output tags: escaped and non-escaped. The only difference
is that escaped tags will have < and > characters HTML-escaped.

```go
<%= "<script>alert('this will be escaped')</script>" %>
<%! "<script>alert('this won't be escaped')</script>" %>
```

## Variables
Gerb attempts to behave as close to Go as possible. The biggest difference is that
method calls and field access is case-insensitive. Gerb supports:

* ints
* float64
* strings
* byte
* fields
* methods
* arrays
* basic operations

For example, the following works:

```go
<%= user.Analysis(count)[start+7:] %>
```

+, -, /, * and % are the only support operations. Currently (and sadly) order of
precedence is left to right and parenthesis cannot be used (parenthesis can be
used in if/elseif statements).

Gerb might occasionally be a little less strict than Go with type conversions,
but not by much.

## Builtins
Go's builtins aren't natively available. However, custom builtin functions can
be registered. The custom buitlin `len` comes pre-registered and behaves much
like the real `len` builtin. You can register your own builtin:

```go
import "github.com/karlseguin/gerb/core"
func init() {
  RegisterBuiltin("add", func(a, b int) int {
    return a + b
  })
}
```

`RegisterBuiltin` isn't threadsafe. It's expected that you'll register your builtins
at startup and then leave it alone.

## Aliases and Package Functions
In addition to custom builtins, it is possible to alias package functions. This
makes functions such as `strings.ToUpper` available to use.

By default, many functions from the `strings` package in addition to
`fmt.Sprintf` and `strconv.Atoi` are available.

Since Go doesn't make it possible to automatically reflect functions exposed from
a package, registration must manually be done on a per-method basis. Like
builtins, this process is not thread safe:

```go
func init() {
  core.RegisterAliases("strings",
    "ToUpper", strings.ToUpper,
    "ToLower", strings.ToLower,
    ...
  )
  core.RegisterAliases("fmt",
    "Sprintf", fmt.Sprintf
  )
}
```

You can see a list of what's currently aliased by looking at
<https://github.com/karlseguin/gerb/blob/master/core/aliases.go>

## Multiple Return Values
If you call a function which returns multiple values, only the first value is
considered/returned. The exception to this rule is with any assignment.

## Assignments
It's possible to create/assign to new or existing variables within a template:

```go
<% name := "leto" %>
```

Assignment supports multiple values, either via an explicit list of values or
a function which returns multiple values:

```go
<% name, power := "goku", 9000 %>
<% name, power := sayan.Stats() %>
```

In fact, you can (but probably shouldn't) mix the two:

```go
<% name, power := sayan.Name(), 9000 %>
```

This is also true for assignments within an `if` or `else if` tag:

```go
<% if n, err := strconv.Atoi(value); err != nil { %>
  The number is <%= n %>
<% } %>
```

## if/elseif/else
If tag closely mimic Go's if statements. Assignments within an if/elseif
is allowed and braces are mandatory.

## for
For tag supports ranged iteration (over slices, arrays, maps and strings)
as well as the traditional C-style (including an empty for `for { ... }`).

The `continue` and `break` tag work as expected within a for loop.

## ++, --, += and -=
There's limited support for these four operators. As a general rule, they should
only be used on simple values (support was added to support the i++ in a `for`
loop).

Here's a couple examples of what **is not supported**:

```go
<% user.PowerLevel++ %>
<% ranks[4]++ %>
```

Put differently, these 4 operators should only ever be used as such:

```go
<% counter++ %>
```

## Newlines
Use `<%%` to trim newlines from literals before a code block. Use `%%>` to trim newlines from a literal following a codeblock.

## Comments
`<%` can be used for comments within templates:

```go
<%# my comment %>
```

Newlines can be trimmed around comments using `<%%#` and `%%>`

## Errors and Logs
`Render` should never fail. By default, `Render` will log errors to stdout. This
behavior can be changed. To disable all logging:

```go
gerb.Configure().Logger(nil)
```

Alternatively, to use your own logger:

```go
gerb.Configure().Logger(new(MyLogger))
```

Your logger must implement `Error(v ...interface{})`.

## Template Inheritance
The `Parse`, `ParseString` and `ParseFile` methods accept a variable length
of parameters. The purpose of this is to support template inheritance:

```go
t := gerb.ParseFile(true, "update.gerb", "member.gerb", "main.gerb")
```

Templates should be specified from innermost to outermost.

The `<% capture NAME {%>...<% } %>` and builtin `yield` can be used to manage
content:

```go
layout := `<title><%= yield("title")%></title> <%= yield %> <footer>...</footer>`
about := `<% content "title" { %>about us<% } %> We are ...`
t, _ := gerb.ParseString(true, about, layout)
```

## Configuration
Gerb can be configured via the fluent interface exposed by `gerb.Configure()`.

- `Logger`: specify the logger to use when an error is encountered rendering
(defaults to stdout)
