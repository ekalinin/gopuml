// SPDX-License-Identifier: MIT
// Provenance: plantuml-mit 1.2026.9 net/sourceforge/plantuml/brotli/Dictionary.java
// Derived from PlantUML, (C) Copyright 2009-2024, Arnaud Roques, MIT License.

package brotli

import "sync"

// Collection of static dictionary words. The content is unpacked from the
// strings of dictionary_data.go when getDictionaryData is called for the
// first time, which saves memory and CPU when the dictionary is not
// required. Upstream's public setData, which lets a caller replace the
// dictionary, is not ported.
var (
	dictionaryOnce sync.Once
	dictionaryData []byte
)

// getDictionaryData is upstream's Dictionary.getData, with a sync.Once in
// place of Java's double-checked locking around DictionaryData's static
// initializer.
func getDictionaryData() []byte {
	dictionaryOnce.Do(func() {
		dictionary := make([]byte, dictionarySize)
		unpackDictionaryData(dictionary, data0, data1, skipFlip)
		dictionaryData = dictionary
	})
	if dictionaryData == nil {
		throw("brotli dictionary is not set")
	}
	return dictionaryData
}
