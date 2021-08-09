// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package whatsnew_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jbowes/whatsnew"
	"github.com/jbowes/whatsnew/impl"
)

func Example() {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Slug:    "you/your-app",
		Cache:   "testdata/update-cache.json",
		Version: "v0.0.1",
	})

	// Run your CLI code and whatnot

	if v, _ := fut.Get(); v != "" {
		fmt.Printf("new release available: %s\n", v)
	}

	// Output:
	// new release available: v0.2.0
}

// This example test isn't really needed, but it keeps the file
// from being an example program, so we can replace the http
// transport etc.
func Example_customFrequency() {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Slug:      "you/your-app",
		Cache:     "testdata/update-cache.json",
		Version:   "0.0.1",
		Frequency: 24 * time.Hour,
	})

	// Run your CLI code and whatnot

	if v, _ := fut.Get(); v != "" {
		fmt.Printf("new release available: %s\n", v)
	}

	// Output:
	// new release available: 0.30.0
}

func init() {
	ctx := context.Background()

	// setup. write out a cache that will hit for one test and
	// miss on the other
	cache := impl.FileCacher{Path: "testdata/update-cache.json"}
	_ = cache.Set(ctx, &impl.Info{
		Version:   "v0.2.0",
		CheckTime: time.Now().Add(-25 * time.Hour),
		Etag:      "whatever",
	})

	// replace http default transport.
	http.DefaultTransport = http.NewFileTransport(
		http.Dir("testdata/example"),
	)
}
