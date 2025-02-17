// Copyright 2022 Molecula Corp. (DBA FeatureBase).
// SPDX-License-Identifier: Apache-2.0
package pilosa

import (
	"testing"

	"github.com/featurebasedb/featurebase/v3/testhook"
)

// mustOpenIndex returns a new, opened index at a temporary path. Panic on error.
func mustOpenIndex(tb testing.TB, opt IndexOptions) *Index {
	path, err := testhook.TempDirInDir(tb, *TempDir, "pilosa-index-")
	if err != nil {
		panic(err)
	}
	h := NewHolder(path, mustHolderConfig())
	index, err := h.CreateIndex("i", opt)
	testhook.Cleanup(tb, func() {
		h.Close()
	})

	if err != nil {
		panic(err)
	}

	index.keys = opt.Keys
	index.trackExistence = opt.TrackExistence

	return index
}
