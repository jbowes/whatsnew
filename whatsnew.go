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
	"errors"
	"fmt"
	"time"

	"github.com/jbowes/semver"

	"github.com/jbowes/whatsnew/impl"
)

// ErrMisconfiguredOptions is returned when incompatible options are set.
// You must use `errors.Is` to check for this error.
var ErrMisconfiguredOptions = errors.New("invalid options provided")

// DefaultFrequency is the default frequency used to check for new
// releases if no override is given. It is one week.
const DefaultFrequency = 7 * 24 * time.Hour

// Timeout values used as options to Check, controlling how long the
// Check is allowed to run.
const (
	DefaultTimeout = 5 * time.Second
	NoTimeout      = time.Duration(-1)
)

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

	// Optional. Controls how often to run a release check.
	// If not provided, DefaultFrequency is used.
	Frequency time.Duration

	// Optional. Sets a maximum duration to run the check before
	// timing out and returning either the cached value, or no update.
	// If not provided, DefaultTimeout is used. Set to NoTimeout (-1)
	// to disable the timeout. Context cancelation is honored, so you
	// may further restrict the deadline with the provided context.
	Timeout time.Duration

	// Slots to override cacher and Releaser
	Cacher   impl.Cacher   // If provided, Cache is ignored.
	Releaser impl.Releaser // If provided, Slug is ignored.
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
	if o.Cacher != nil && o.Cache != "" {
		return fmt.Errorf("cache and cacher set: %w", ErrMisconfiguredOptions)
	}

	if o.Releaser != nil && o.Slug != "" {
		return fmt.Errorf("releaser and slug set: %w", ErrMisconfiguredOptions)
	}

	if o.Cacher == nil {
		o.Cacher = &impl.FileCacher{Path: o.Cache}
	}

	if o.Releaser == nil {
		o.Releaser = &impl.GitHubReleaser{URL: fmt.Sprintf("https://api.github.com/repos/%s/releases", o.Slug)}
	}

	if o.Frequency == 0 {
		o.Frequency = DefaultFrequency
	}

	if o.Timeout == 0 {
		o.Timeout = DefaultTimeout
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

	if opts.Timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	i, err := opts.Cacher.Get(ctx)
	if err != nil {
		i = &impl.Info{}
	}
	iHasV, iVer, _ := parseV(i.Version)

	now := time.Now()

	optHasV, optVer, _ := parseV(opts.Version)
	nextVer := optVer
	nextHasV := optHasV
	if now.Sub(i.CheckTime) < opts.Frequency {
		nextVer = iVer
		nextHasV = iHasV
	} else {
		rels, etag, err := opts.Releaser.Get(ctx, i.Etag)
		if err != nil {
			// If we error, fall back to possibly using the value from the store
			nextVer = iVer
			nextHasV = iHasV
		} else if len(rels) == 0 {
			// Cached result. refresh the checktime and store.
			_ = opts.Cacher.Set(ctx, &impl.Info{
				CheckTime: now,
				Etag:      etag,
				Version:   i.Version,
			})

			nextVer = iVer
			nextHasV = iHasV
		} else {
			// find the biggest non-prerelease version in releases.
			// TODO: could look at more than the first page. would only matter
			// for concurrent patch releases etc.
			var newVer *semver.Version
			newHasV := ""
			for _, rel := range rels {
				hv, pv, err := parseV(rel.TagName)
				switch {
				case err != nil: // not a valid semver tag
				case rel.Draft:
				case rel.Prerelease || pv.Prerelease() != "":
				case newVer.Compare(pv) < 0:
					newVer = pv
					newHasV = hv
				}
			}

			// TODO: make sure newVer is set
			_ = opts.Cacher.Set(ctx, &impl.Info{
				CheckTime: now,
				Etag:      etag,
				Version:   newHasV + newVer.String(), // we store the latest from the remote ignoring what's installed.
			})

			if nextVer.Compare(newVer) < 1 {
				nextVer = newVer
				nextHasV = newHasV
			}
		}
	}

	if optVer.Compare(nextVer) >= 0 {
		return "", nil
	}

	return nextHasV + nextVer.String(), nil
}

func parseV(s string) (string, *semver.Version, error) {
	// TODO: parsing out the v and holding it isn't great.
	hasV := ""
	if s != "" && s[0] == 'v' {
		hasV = "v"
		s = s[1:]
	}

	v, err := semver.Parse(s)

	return hasV, v, err
}
