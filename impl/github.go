// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GitHubReleaser is the default Releaser used in whatsnew.
type GitHubReleaser struct {
	URL    string       // a complete URL to the releases API.
	Client *http.Client // if not set, http.DefaultClient is used.
}

// Get a list of releases.
func (g *GitHubReleaser) Get(ctx context.Context, etag string) ([]Release, string, error) {
	req, err := http.NewRequest(http.MethodGet, g.URL, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	req = req.WithContext(ctx)

	c := g.Client
	if c == nil {
		c = http.DefaultClient
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if etag != "" && resp.StatusCode == http.StatusNotModified {
		return nil, etag, nil // this will fall back to existing stuff.
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("error getting updates: %s", resp.Status)
	}

	var rels []Release
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&rels); err != nil {
		return nil, "", err
	}

	return rels, resp.Header.Get("Etag"), nil
}
