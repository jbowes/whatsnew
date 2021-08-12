<!--
  Attractive html formatting for rendering in github. sorry text editor
  readers! Besides the header and section links, everything should be clean and
  readable.
-->
<h1 align="center">whatsnew</h1>
<p align="center"><i>Check for new github releases of your Golang application 🎊</i></p>

<div align="center">
  <a href="https://pkg.go.dev/github.com/jbowes/whatsnew"><img src="https://pkg.go.dev/badge/github.com/jbowes/whatsnew.svg" alt="Go Reference"></a>
  <img alt="Enterprise Ready" src="https://img.shields.io/badge/Enterprise-Ready-pink.svg" >
  <a href="https://github.com/jbowes/whatsnew/releases/latest"><img alt="GitHub tag" src="https://img.shields.io/github/tag/jbowes/whatsnew.svg"></a>
  <a href="https://github.com/jbowes/whatsnew/actions/workflows/go.yml"><img alt="Build Status" src="https://github.com/jbowes/whatsnew/actions/workflows/go.yml/badge.svg?branch=main"></a>
  <a href="./LICENSE"><img alt="BSD license" src="https://img.shields.io/badge/license-BSD-blue.svg"></a>
  <a href="https://codecov.io/gh/jbowes/whatsnew"><img alt="codecov" src="https://img.shields.io/codecov/c/github/jbowes/whatsnew.svg"></a>
  <a href="https://goreportcard.com/report/github.com/jbowes/whatsnew"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/jbowes/whatsnew"></a>
</div><br /><br />

---

[`whatsnew`][godoc] provides a simple way to check GitHub for new releases of
your Go application. It saves results between runs, uses etags to speed up responses, and tries to minimize the overhead it adds to an otherwise fast
application CLI run.

If caching to disk, or reading from GitHub don't work for you, you can
customize the behaviour.

## Quick start

```go
import (
  "context"
  "github.com/jbowes/whatsnew"
)

func main() {
	ctx := context.Background()

	// Start a whatsnew Check
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Slug:    "you/your-app",
		Cache:   "testdata/update-cache.json",
		Version: "v0.0.1",
	})

	// Run your CLI code and whatnot

	// Wait for the Check to complete, and show the results
	if v, _ := fut.Get(); v != "" {
		fmt.Printf("new release available: %s\n", v)
	}
}
```

For more usage and examples, see the [GoDoc Reference][godoc]

## Alternatives

`whatsnew` only **checks** for releases. If you're looking for a package that
will let your application **update itself**, or you prefer packages that start
with `go-`, consider one of these:
- [go-github-selfupdate](https://github.com/rhysd/go-github-selfupdate)
- [go-selfupdate](https://github.com/sanbornm/go-selfupdate)
- [go-update](https://github.com/inconshreveable/go-update)

## Contributing

I would love your help!

`whatsnew` is still a work in progress. You can help by:

- Opening a pull request to resolve an [open issue][issues].
- Adding a feature or enhancement of your own! If it might be big, please
  [open an issue][enhancement] first so we can discuss it.
- Improving this `README` or adding other documentation to `whatsnew`.
- Letting [me] know if you're using `whatsnew`.

[godoc]: https://pkg.go.dev/github.com/jbowes/whatsnew

[issues]: ./issues
[bug]: ./issues/new?labels=bug
[enhancement]: ./issues/new?labels=enhancement

[me]: https://twitter.com/jrbowes
