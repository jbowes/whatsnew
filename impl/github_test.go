// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/jbowes/whatsnew/impl"
)

func TestGihubReleaser(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		URL: "http://github.com/repos/you/your-app/releases",
		Client: &http.Client{
			Transport: http.NewFileTransport(
				http.Dir("../testdata/example"),
			),
		},
	}
	rels, etag, err := ghr.Get(ctx, `"some-etag"`)
	if err != nil {
		t.Errorf("got unexpected error: %s", err)
	}

	if len(rels) != 1 {
		t.Errorf("wrong number of releases. expected: %d got: %d", 1, len(rels))
	}
	if rels[0].TagName != "0.30.0" {
		t.Errorf("wrong tag name. expected: %s got: %s", "0.30.0", rels[0].TagName)
	}

	if etag != "" {
		t.Errorf("wrong etag. expected: %s got: %s", "", etag)
	}
}

func TestGihubReleaser_errorOn404(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		URL: "http://github.com/repos/you/your-app/badurl",
		Client: &http.Client{
			Transport: http.NewFileTransport(
				http.Dir("../testdata/example"),
			),
		},
	}
	_, _, err := ghr.Get(ctx, `"some-etag"`)
	if err == nil {
		t.Error("expected error but got none")
	}
}

type errTripper struct{}

func (errTripper) RoundTrip(*http.Request) (*http.Response, error) { return nil, errors.New("oops") }

func TestGihubReleaser_errorOnRequest(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		URL: "http://github.com/repos/you/your-app/badurl",
		Client: &http.Client{
			Transport: errTripper{},
		},
	}
	_, _, err := ghr.Get(ctx, `"some-etag"`)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestGihubReleaser_errorOnBadURL(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		// This test case is really stretching it, but those coverage
		// metrics are just so enticing
		URL: "\x7f",
	}
	_, _, err := ghr.Get(ctx, `"some-etag"`)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestGihubReleaser_errorOnBadJSON(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		URL: "http://github.com/repos/you/your-app/bad-json",
		Client: &http.Client{
			Transport: http.NewFileTransport(
				http.Dir("../testdata/example"),
			),
		},
	}
	_, _, err := ghr.Get(ctx, `"some-etag"`)
	if err == nil {
		t.Error("expected error but got none")
	}
}

type etagTransport struct{}

func (etagTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Header.Get("If-None-Match") != `"some-etag"` {
		return nil, errors.New("expected etag")
	}

	resp := &http.Response{
		StatusCode: http.StatusNotModified,
	}
	return resp, nil
}

func TestGihubReleaser_supportsEtag(t *testing.T) {
	ctx := context.Background()
	ghr := &impl.GitHubReleaser{
		URL: "http://github.com/repos/you/your-app/releases",
		Client: &http.Client{
			Transport: etagTransport{},
		},
	}
	etag := `"some-etag"`
	rels, outEtag, err := ghr.Get(ctx, etag)
	if err != nil {
		t.Error("unexpected error:", err)
	}
	if len(rels) != 0 {
		t.Error("expected no rels but got some")
	}
	if outEtag != etag {
		t.Error("incorrect etag. wanted:", etag, "got:", outEtag)
	}
}
