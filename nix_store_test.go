package nix

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const testDerivationJSON = `{
  "name": "go-bindings-test",
  "version": 4,
  "outputs": {
    "out": {
      "path": "d1nf171c83f8aczqcn20r20r1bisij3i-go-bindings-test"
    }
  },
  "inputs": {
    "srcs": [],
    "drvs": {}
  },
  "system": "x86_64-linux",
  "builder": "/bin/sh",
  "args": [],
  "env": {
    "builder": "/bin/sh",
    "name": "go-bindings-test",
    "out": "/nix/store/d1nf171c83f8aczqcn20r20r1bisij3i-go-bindings-test",
    "system": "x86_64-linux"
  }
}`

func errMsgString(t *testing.T, ctx *NixCContext) string {
	t.Helper()

	msg := ErrMsg(nil, ctx)
	if msg == nil {
		return ""
	}
	return ownedCString(t, msg)
}

func newTestStore(t *testing.T, ctx *NixCContext) *Store {
	t.Helper()

	if got := LibstoreInitNoLoadConfig(ctx); got != NixOk {
		t.Fatalf("LibstoreInitNoLoadConfig = %v, want %v: %s", got, NixOk, errMsgString(t, ctx))
	}

	store := StoreOpen(ctx, "dummy://", StoreParams{})
	if store == nil {
		t.Fatalf("StoreOpen(dummy://) returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		StoreFree(store)
	})

	return store
}

func mustParseJSON(t *testing.T, raw string) map[string]any {
	t.Helper()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("json.Unmarshal exported derivation: %v\n%s", err, raw)
	}

	return parsed
}

func storePathNameString(t *testing.T, path *StorePath) string {
	t.Helper()

	return ownedCString(t, StorePathName(path))
}

func storePathHashBytes(t *testing.T, ctx *NixCContext, path *StorePath) [20]byte {
	t.Helper()

	var hash StorePathHashPart
	if got := StorePathHash(ctx, path, &hash); got != NixOk {
		hash.Free()
		t.Fatalf("StorePathHash = %v, want %v: %s", got, NixOk, errMsgString(t, ctx))
	}
	hash.Deref()
	hash.Free()

	return hash.Bytes
}

func TestNixStoreOpenAndStrings(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	storeDir := ownedCString(t, StoreGetStoredir(ctx, store))
	if storeDir != "/nix/store" {
		t.Fatalf("StoreGetStoredir = %q, want /nix/store", storeDir)
	}

	uri := ownedCString(t, StoreGetUri(ctx, store))
	if strings.TrimSpace(uri) == "" {
		t.Fatal("StoreGetUri returned an empty string")
	}

	version := StoreGetVersion(ctx, store)
	if version != nil {
		StringFree(version)
	}
}

func TestNixStoreParsePathRealPathAndValidity(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	const rawPath = "/nix/store/00000000000000000000000000000000-demo"
	path := StoreParsePath(ctx, store, rawPath)
	if path == nil {
		t.Fatalf("StoreParsePath returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		StorePathFree(path)
	})

	if name := storePathNameString(t, path); name != "demo" {
		t.Fatalf("StorePathName = %q, want demo", name)
	}

	zeroHash := storePathHashBytes(t, ctx, path)
	if zeroHash != ([20]byte{}) {
		t.Fatalf("StorePathHash for zero path = %v, want all zero bytes", zeroHash)
	}

	realPath := ownedCString(t, StoreRealPath(ctx, store, path))
	if realPath != rawPath {
		t.Fatalf("StoreRealPath = %q, want %q", realPath, rawPath)
	}

	if StoreIsValidPath(ctx, store, path) {
		t.Fatal("dummy store unexpectedly reports arbitrary path valid")
	}

	clone := StorePathClone(path)
	if clone == nil {
		t.Fatalf("StorePathClone returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		StorePathFree(clone)
	})
	if name := storePathNameString(t, clone); name != "demo" {
		t.Fatalf("StorePathName(clone) = %q, want demo", name)
	}
	if cloneHash := storePathHashBytes(t, ctx, clone); cloneHash != zeroHash {
		t.Fatalf("StorePathHash(clone) = %v, want %v", cloneHash, zeroHash)
	}

	const nonZeroRawPath = "/nix/store/11111111111111111111111111111111-demo"
	nonZeroPath := StoreParsePath(ctx, store, nonZeroRawPath)
	if nonZeroPath == nil {
		t.Fatalf("StoreParsePath(non-zero) returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		StorePathFree(nonZeroPath)
	})

	nonZeroHash := StorePathHashPart{
		Bytes: storePathHashBytes(t, ctx, nonZeroPath),
	}
	if nonZeroHash.Bytes == ([20]byte{}) {
		t.Fatal("StorePathHash for non-zero path returned all zero bytes")
	}

	const createdName = "created-from-parts"
	created := StoreCreateFromParts(ctx, &nonZeroHash, createdName, uint64(len(createdName)))
	nonZeroHash.Free()
	if created == nil {
		t.Fatalf("StoreCreateFromParts returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		StorePathFree(created)
	})
	if name := storePathNameString(t, created); name != createdName {
		t.Fatalf("StorePathName(created) = %q, want %q", name, createdName)
	}
	if createdHash := storePathHashBytes(t, ctx, created); createdHash != nonZeroHash.Bytes {
		t.Fatalf("StorePathHash(created) = %v, want %v", createdHash, nonZeroHash.Bytes)
	}
}

func TestNixStoreParseInvalidPathSetsContextError(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	path := StoreParsePath(ctx, store, "/not-a-store-path")
	if path != nil {
		t.Fatal("StoreParsePath unexpectedly accepted invalid path")
	}
	if got := ErrCode(ctx); got == NixOk {
		t.Fatalf("ErrCode after invalid StoreParsePath = %v, want non-OK", got)
	}
	if msg := errMsgString(t, ctx); strings.TrimSpace(msg) == "" {
		t.Fatal("ErrMsg after invalid StoreParsePath returned an empty string")
	}

	ClearErr(ctx)
	if got := ErrCode(ctx); got != NixOk {
		t.Fatalf("ErrCode after ClearErr = %v, want %v", got, NixOk)
	}
}

func TestNixStoreOpenInvalidUriSetsContextError(t *testing.T) {
	ctx := newTestContext(t)
	if got := LibstoreInitNoLoadConfig(ctx); got != NixOk {
		t.Fatalf("LibstoreInitNoLoadConfig = %v, want %v: %s", got, NixOk, errMsgString(t, ctx))
	}

	store := StoreOpen(ctx, "go-bindings-test-invalid-store://", StoreParams{})
	if store != nil {
		StoreFree(store)
		t.Fatal("StoreOpen unexpectedly accepted invalid URI")
	}
	if got := ErrCode(ctx); got == NixOk {
		t.Fatalf("ErrCode after invalid StoreOpen = %v, want non-OK", got)
	}
	if msg := errMsgString(t, ctx); strings.TrimSpace(msg) == "" {
		t.Fatal("ErrMsg after invalid StoreOpen returned an empty string")
	}
}

func TestNixStoreDerivationLifecycleAndJSON(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	derivation := DerivationFromJson(ctx, store, testDerivationJSON)
	if derivation == nil {
		t.Fatalf("DerivationFromJson returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		DerivationFree(derivation)
	})

	exported := ownedCString(t, DerivationToJson(ctx, derivation))
	if strings.TrimSpace(exported) == "" {
		t.Fatal("DerivationToJson returned an empty string")
	}

	exportedJSON := mustParseJSON(t, exported)
	if exportedJSON["name"] != "go-bindings-test" {
		t.Fatalf("exported derivation name = %v, want go-bindings-test", exportedJSON["name"])
	}
	if exportedJSON["version"] != float64(4) {
		t.Fatalf("exported derivation version = %v, want 4", exportedJSON["version"])
	}
	outputs, ok := exportedJSON["outputs"].(map[string]any)
	if !ok {
		t.Fatalf("exported derivation outputs = %T, want JSON object", exportedJSON["outputs"])
	}
	if _, ok := outputs["out"]; !ok {
		t.Fatal("exported derivation outputs missing out")
	}

	clone := DerivationClone(derivation)
	if clone == nil {
		t.Fatalf("DerivationClone returned nil: err=%v msg=%q", ErrCode(ctx), errMsgString(t, ctx))
	}
	t.Cleanup(func() {
		DerivationFree(clone)
	})

	cloneExported := ownedCString(t, DerivationToJson(ctx, clone))
	cloneJSON := mustParseJSON(t, cloneExported)
	if !reflect.DeepEqual(cloneJSON, exportedJSON) {
		t.Fatalf("DerivationToJson(clone) = %#v, want %#v", cloneJSON, exportedJSON)
	}
}

func TestNixStoreInvalidDerivationJSONSetsContextError(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	derivation := DerivationFromJson(ctx, store, `{"version": 4}`)
	if derivation != nil {
		DerivationFree(derivation)
		t.Fatal("DerivationFromJson unexpectedly accepted invalid derivation JSON")
	}
	if got := ErrCode(ctx); got == NixOk {
		t.Fatalf("ErrCode after invalid DerivationFromJson = %v, want non-OK", got)
	}
	if msg := errMsgString(t, ctx); strings.TrimSpace(msg) == "" {
		t.Fatal("ErrMsg after invalid DerivationFromJson returned an empty string")
	}

	ClearErr(ctx)
	if got := ErrCode(ctx); got != NixOk {
		t.Fatalf("ErrCode after ClearErr = %v, want %v", got, NixOk)
	}
}

func TestNixStoreCopyPathRejectsNilArguments(t *testing.T) {
	ctx := newTestContext(t)
	store := newTestStore(t, ctx)

	if got := StoreCopyPath(ctx, nil, store, nil, false, false); got == NixOk {
		t.Fatalf("StoreCopyPath with nil arguments = %v, want non-OK", got)
	}
	if msg := errMsgString(t, ctx); strings.TrimSpace(msg) == "" {
		t.Fatal("ErrMsg after invalid StoreCopyPath returned an empty string")
	}
}
