// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package whatsnew checks for new GitHub releases of your Golang application.
//
// whatsnew expects that versions follow https://semver.org
//
// By default, whatsnew saves cached results to disk, and retrieves
// known releases from public GitHub repos. If you need to modify this
// behaviour, see the `impl` subpackage for details on how to provide
// your own impl.Cacher or impl.Releaser.
package whatsnew

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/mod/semver"

	"github.com/jbowes/whatsnew/impl"
)

// DefaultFrequency is the default frequency used to check for new
// releases if no override is given. It is one week.
const DefaultFrequency = 7 * 24 * time.Hour

type result struct {
	v   string
	err error
}

// Future holds the future results from a call to Check.
type Future struct {
	c <-chan *result
	r *result
}

// Get returns the results from a call to Check. Check runs in its own
// goroutine; Get will block waiting for the goroutine to complete.
//
// If an updated version is detected, that version string is returned.
// If no update is found, the empty string is returned.
func (f *Future) Get() (string, error) {
	if f.r == nil {
		f.r = <-f.c
	}

	return f.r.v, f.r.err
}

// Options sets both required and optional values for running a Check.
type Options struct {
	Slug    string // The GitHub repository slug, eg `jbowes/whatsnew`
	Cache   string // A full file path to store the cache. Should end in `.json`
	Version string // The current semver version of the program to check.

	// Optional. if not provided DefaultFrequency is used.
	Frequency time.Duration

	// Slots to override cacher and Releaser
	// TODO: make these public
	cacher   impl.Cacher
	releaser impl.Releaser
}

/* TODO: prerelease flags
 Flags to modify prelease etc behaviour
	Flags Flag


type Flag byte

const (
	NoFlags        Flag = 0
	SamePrerelease Flag = 1 << iota
	AcrossPrerelease
	IntoPrerelease
)
*/

func (o *Options) resolve() error {
	// TODO: don't overwrite if provided.
	// TODO: error if strings + impls are provided.
	o.cacher = &impl.FileCacher{Path: o.Cache}
	o.releaser = &impl.GitHubReleaser{URL: fmt.Sprintf("https://api.github.com/repos/%s/releases", o.Slug)}

	if o.Frequency == 0 {
		o.Frequency = DefaultFrequency
	}

	return nil
}

// Check checks github for a newer release of the configured application.
// Check is primarily meant to be run in short-lived CLI applications,
// and should be called before you do your application's main work.
//
// It returns a Future. After your application's main work is done,
// call Get() on the future to get the result and error.
func Check(ctx context.Context, opts *Options) *Future {
	c := make(chan *result)
	f := Future{c: c}

	go func() {
		r := result{}
		r.v, r.err = doWork(ctx, opts)
		c <- &r
	}()

	return &f
}

// TODO: return if this is a new check or not? could be useful for less spammy
// update notice.
func doWork(ctx context.Context, opts *Options) (string, error) {
	if err := opts.resolve(); err != nil {
		return "", err
	}

	i, err := opts.cacher.Get(ctx)
	if err != nil {
		i = &impl.Info{}
	}

	now := time.Now()

	nextVer := opts.Version
	if now.Sub(i.CheckTime) < opts.Frequency {
		if cmp(nextVer, i.Version) < 0 {
			nextVer = i.Version
		}
	} else {
		rels, etag, err := opts.releaser.Get(ctx, i.Etag)
		// If we error, fall back to possibly using the value from the store
		if err != nil || len(rels) == 0 {
			if cmp(nextVer, i.Version) < 0 {
				return i.Version, nil
			}
		}
		// find the biggest non-prerelease version in releases.
		// TODO: could look at more than the first page. would only matter
		// for concurrent patch releases etc.
		var newVer string
		for _, rel := range rels {
			switch {
			case rel.Draft:
			case !isValid(rel.TagName):
			case rel.Prerelease || isPrerelease(rel.TagName):
			case cmp(newVer, rel.TagName) < 0:
				newVer = rel.TagName
			}
		}

		if cmp(nextVer, newVer) < 1 {
			nextVer = newVer
		}

		_ = opts.cacher.Set(ctx, &impl.Info{
			CheckTime: now,
			Etag:      etag,
			Version:   newVer, // we store the latest from the remote ignoring whats installed.
		})
	}

	if cmp(opts.Version, nextVer) == 0 {
		return "", nil
	}

	return nextVer, nil
}

// cmp is like x/mod/semver except it allows versions to not start with a v.
func cmp(v1, v2 string) int {
	v1 = maybeV(v1)
	v2 = maybeV(v2)

	return semver.Compare(v1, v2)
}

func isValid(v1 string) bool      { return semver.IsValid(maybeV(v1)) }
func isPrerelease(v1 string) bool { return semver.Prerelease(maybeV(v1)) != "" }

func maybeV(v string) string {
	if v != "" && v[0] != 'v' {
		return "v" + v
	}

	return v
}
