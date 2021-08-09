// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package impl provides interfaces for implementing custom whatsnew caches or
// release fetchers.
package impl

import (
	"context"
	"time"
)

// Cacher sets and gets cached Info for release checks.
//
// Implement a Cacher to change where and how whatsnew persists
// previous release checks, or disable persistence altogether.
type Cacher interface {
	Get(context.Context) (*Info, error)
	Set(context.Context, *Info) error
}

// Info is cached information about the newest last-seen release.
type Info struct {
	CheckTime time.Time `json:"check_time"` // When the check was last run
	Version   string    `json:"version"`    // The largest/newest version seen in the last check
	Etag      string    `json:"etag"`       // An entity tag to aid in refetchin.
}

// Releaser gets a list of releases from a source.
//
// Implement a Releaser to acess private GitHub repositories, or
// get releases from other hosting services or locations.
type Releaser interface {
	// Get a list of releases.
	//
	// Where possible, it should honor the provided etag, and return a
	// new etag for every call. If the data has not changed based on the
	// provied etag, return the provided etag, and an empty list of releases.
	Get(ctx context.Context, etag string) (releases []Release, newEtag string, err error)
}

// Release is a single release entry from a releaser.
// It is modeled after the fields in GitHub releases.
type Release struct {
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
	TagName    string `json:"tag_name"`
}
