// Copyright (c) 2021 James Bowes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package impl_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbowes/whatsnew/impl"
)

func TestFileCacher_roundTrip(t *testing.T) {
	ctx := context.Background()

	path, err := os.MkdirTemp("", "*")
	if err != nil {
		t.Fatal("couldn't set up temp dir")
	}
	defer os.RemoveAll(path)

	fc := impl.FileCacher{Path: filepath.Join(path, "test-cache.json")}

	now := time.Now()
	version := fmt.Sprintf("v1.1.2+%s", filepath.Base(path)) // random per run name
	etag := filepath.Base(path)
	err = fc.Set(ctx, &impl.Info{
		CheckTime: now,
		Version:   version,
		Etag:      etag,
	})

	if err != nil {
		t.Errorf("error running set: %s", err)
	}

	out, err := fc.Get(ctx)
	if err != nil {
		t.Errorf("error running get: %s", err)
	}

	if !out.CheckTime.Equal(now) {
		t.Errorf("CheckTime wrong. get: %s wanted: %s", out.CheckTime, now)
	}

	if out.Version != version {
		t.Errorf("Version wrong. get: %s wanted: %s", out.Version, version)
	}

	if out.Etag != etag {
		t.Errorf("Etag wrong. get: %s wanted: %s", out.Etag, etag)
	}

}

func TestFileCacher_errOnWrite(t *testing.T) {
	ctx := context.Background()

	fc := impl.FileCacher{Path: filepath.Join("/", "test-cache.json")}

	err := fc.Set(ctx, &impl.Info{
		CheckTime: time.Now(),
		Version:   "v1.1.2",
		Etag:      `"some-etag"`,
	})

	if err == nil {
		t.Errorf("expected err but go none")
	}
}

func TestFileCacher_errOnMkDir(t *testing.T) {
	ctx := context.Background()

	fc := impl.FileCacher{Path: filepath.Join("/", "whatsnew-test", "test-cache.json")}

	err := fc.Set(ctx, &impl.Info{
		CheckTime: time.Now(),
		Version:   "v1.1.2",
		Etag:      `"some-etag"`,
	})

	if err == nil {
		t.Errorf("expected err but go none")
	}
}
