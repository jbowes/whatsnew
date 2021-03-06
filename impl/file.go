// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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
	defer r.Close()

	var i Info
	dec := json.NewDecoder(r)
	err = dec.Decode(&i)
	return &i, err
}

// Set cached release Info.
func (f *FileCacher) Set(_ context.Context, i *Info) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0750); err != nil {
		return err
	}

	w, err := os.Create(f.Path)
	if err != nil {
		return err
	}
	defer w.Close()

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(i)
}
