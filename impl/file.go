// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl

import (
	"context"
	"encoding/json"
	"os"
)

// FileCacher is the default Cacher used in whatsnew.
type FileCacher struct {
	Path string
}

// Get cached release Info.
func (f *FileCacher) Get(context.Context) (*Info, error) {
	r, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}

	var i Info
	dec := json.NewDecoder(r)
	err = dec.Decode(&i)
	return &i, err
}

// Set cached release Info.
func (f *FileCacher) Set(_ context.Context, i *Info) error {
	w, err := os.Create(f.Path)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(i)
}
