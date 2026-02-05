## Katti (കത്തി)
Katti is a PEG-inspired parser combinator library for Go. Parsers are constructed by composing matcher functions rather than writing a grammar DSL. The library provides a small set of core combinators, and additional combinators can be implemented by the `Matcher` type.

Parsing is driven by error signaling: a matcher either consumes input by mutating a `MatchResult` or fails by returning `ErrNoMatch`. Input consumption is explicit through the `Match` and `Rest` fields of `MatchResult`.

For practical examples and usage patterns, check out the [examples/](examples/) directory in this repository.
