package nix

import (
	"strings"
	"testing"
)

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
