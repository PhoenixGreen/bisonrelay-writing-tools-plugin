// Command bisonrelay-spellcheck-plugin is a Bison Relay dynamic-wasm plugin
// providing the spellcheck-data capability: a dictionary and a set of
// writing-style rules that Bison Relay's own composer UI applies to chat,
// post and comment text as it is typed.
//
// It is headless -- no nav item, no screens -- so beyond the `alloc` export
// every plugin must provide, it implements only `get_spellcheck_data`. There
// is no fetching and no per-keystroke work here: the data is static, handed
// over once when the plugin is enabled, and every match after that happens
// in the app.
//
// That also means the plugin never sees a word anyone types. It has no
// network and no filesystem access, and needs neither.
//
// This file is only the WebAssembly plumbing; the content lives in package
// spellcheck. Build with build.sh (GOOS=wasip1 GOARCH=wasm go build
// -buildmode=c-shared); see wasmhost's package doc in the bisonrelay repo
// for the ABI it implements.
package main

import (
	_ "embed"
	"encoding/json"
	"unsafe"

	"github.com/PhoenixGreen/bisonrelay-spellcheck-plugin/spellcheck"
)

//go:embed words.txt
var wordsTXT string

// --- ABI plumbing every dynamic-wasm plugin needs ---

// pinned keeps allocated buffers alive: nothing else references them once
// alloc returns, so without this the Go GC could collect the memory the host
// is about to write to (or has just written to) before it is read.
var pinned = map[int32][]byte{}

//go:wasmexport alloc
func alloc(size int32) int32 {
	if size == 0 {
		size = 1
	}
	buf := make([]byte, size)
	// go vet flags this uintptr(unsafe.Pointer(...)) conversion as a
	// possible misuse since the address outlives the expression it is taken
	// in -- but Go's allocator doesn't move objects, and buf is kept alive
	// via pinned (keyed by this same address) for as long as the host needs
	// it, so there is no actual staleness risk here.
	ptr := int32(uintptr(unsafe.Pointer(&buf[0])))
	pinned[ptr] = buf
	return ptr
}

func writeResult(data []byte) uint64 {
	ptr := alloc(int32(len(data)))
	copy(pinned[ptr], data)
	return (uint64(uint32(ptr)) << 32) | uint64(len(data))
}

//go:wasmexport get_spellcheck_data
func getSpellcheckData() uint64 {
	b, err := json.Marshal(spellcheck.Data{
		Words:        spellcheck.ParseWords(wordsTXT),
		GrammarRules: spellcheck.Rules,
	})
	if err != nil {
		// Returning zero bytes is how this ABI says "I could not answer";
		// the host logs it and carries on without this provider's data.
		return 0
	}
	return writeResult(b)
}

func main() {}
