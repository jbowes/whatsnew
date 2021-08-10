// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jbowes/whatsnew"
	"github.com/jbowes/whatsnew/impl"
)

// noopCache is a whatsnew Cacher that does nothing, ensuring we always
// check our Releaser, and never try and persist results.
type noopCache struct{}

func (noopCache) Get(context.Context) (*impl.Info, error) { return &impl.Info{}, nil }
func (noopCache) Set(context.Context, *impl.Info) error   { return nil }

func ExampleCacher_noopCache() {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Slug:    "you/your-app",
		Version: "v0.0.1",
		Cacher:  noopCache{}, // Use the custom Cacher
	})

	// Run your CLI code and whatnot

	if v, _ := fut.Get(); v != "" {
		fmt.Printf("new release available: %s\n", v)
	}

	// Output:
	// new release available: 0.30.0
}

// staticReleaser is a whatsnew Releaser that returns a predefined set of
// releases. This isn't useful in practice, outside of tests. But it makes
// for an ok example of how you'd create your own releaser.
type staticReleaser struct{}

func (staticReleaser) Get(context.Context, string) ([]impl.Release, string, error) {
	return []impl.Release{
		{TagName: "v0.99.0"},
	}, "some-etag", nil
}

func ExampleReleaser_staticReleases() {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Cache:    "testdata/update-cache.json",
		Version:  "v0.0.1",
		Releaser: staticReleaser{}, // Use the custom Releaser
	})

	// Run your CLI code and whatnot

	if v, _ := fut.Get(); v != "" {
		fmt.Printf("new release available: %s\n", v)
	}

	// Output:
	// new release available: v0.99.0
}

func init() {
	// replace http default transport.
	http.DefaultTransport = http.NewFileTransport(
		http.Dir("../testdata/example"),
	)
}
