// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package whatsnew_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jbowes/whatsnew"
	"github.com/jbowes/whatsnew/impl"
)

type testCacher struct {
	info *impl.Info
	err  error
}

func (t *testCacher) Get(context.Context) (*impl.Info, error) { return t.info, t.err }
func (testCacher) Set(context.Context, *impl.Info) error      { return nil }

type testReleaser struct {
	releases []impl.Release
	err      error
}

func (t *testReleaser) Get(context.Context, string) ([]impl.Release, string, error) {
	return t.releases, "some-etag", t.err
}

func TestCheck(t *testing.T) {
	ctx := context.Background()
	tcs := map[string]struct {
		releases []impl.Release
		cacheErr error
		out      string
	}{
		"ok": {
			releases: []impl.Release{{TagName: "v1.0.1"}},
			out:      "v1.0.1",
		},
		"no update": {out: ""},
		"draft skipped": {
			releases: []impl.Release{{TagName: "v1.0.1", Draft: true}},
			out:      "",
		},
		"non-semver ignored": {
			releases: []impl.Release{{TagName: "cookies"}},
			out:      "",
		},
		"prerelease skipped": {
			releases: []impl.Release{{TagName: "v1.0.1", Prerelease: true}},
			out:      "",
		},
		"prerelease (semver) skipped": {
			releases: []impl.Release{{TagName: "v1.0.1-beta.2"}},
			out:      "",
		},
		"newest used": {
			releases: []impl.Release{
				{TagName: "v1.0.1"},
				{TagName: "v1.1.1"},
				{TagName: "v1.0.2"},
			},
			out: "v1.1.1",
		},
		"cacher get err": {
			releases: []impl.Release{{TagName: "v1.1.1"}},
			cacheErr: errors.New("oops"),
			out:      "v1.1.1",
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			fut := whatsnew.Check(ctx, &whatsnew.Options{
				Version:  "v1.0.0",
				Cacher:   &testCacher{info: &impl.Info{}, err: tc.cacheErr},
				Releaser: &testReleaser{releases: tc.releases},
			})

			res, err := fut.Get()
			if res != tc.out {
				t.Errorf("versions did not match. got: %s, want: %s", res, tc.out)
			}
			if err != nil {
				t.Errorf("expected nil error. got: %s", err)
			}
		})
	}
}

func TestCheck_fallsBackToCacheOnReleaserError(t *testing.T) {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Version:  "v1.0.0",
		Cacher:   &testCacher{info: &impl.Info{Version: "v1.0.1"}},
		Releaser: &testReleaser{err: errors.New("oops")},
	})

	res, err := fut.Get()
	if res != "v1.0.1" {
		t.Errorf("versions did not match. got: %s, want: %s", res, "v1.0.1")
	}
	if err != nil {
		t.Errorf("expected nil error. got: %s", err)
	}
}

func TestRun_isRepeatable(t *testing.T) {
	ctx := context.Background()
	fut := whatsnew.Check(ctx, &whatsnew.Options{
		Version:  "v1.0.0",
		Cacher:   &testCacher{info: &impl.Info{}},
		Releaser: &testReleaser{releases: []impl.Release{{TagName: "v1.0.1"}}},
	})

	res, err := fut.Get()
	if res != "v1.0.1" {
		t.Errorf("versions did not match. got: %s, want: %s", res, "v1.0.1")
	}
	if err != nil {
		t.Errorf("expected nil error. got: %s", err)
	}

	res2, err2 := fut.Get()
	if res != res2 {
		t.Errorf("repeated versions did not match. got: %s, want: %s", res2, res)
	}
	if err != nil {
		t.Errorf("repeated error not nil. got: %s", err2)
	}
}
