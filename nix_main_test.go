package nix

import (
	"strings"
	"testing"
)

func newTestMainContext(t *testing.T) *NixCContext {
	t.Helper()

	ctx := newTestContext(t)
	if got := LibutilInit(ctx); got != NixOk {
		t.Fatalf("LibutilInit = %v, want %v: %s", got, NixOk, errMsgString(t, ctx))
	}

	return ctx
}

func TestNixMainSetLogFormat(t *testing.T) {
	ctx := newTestMainContext(t)

	formats := []string{
		"raw",
		"raw-with-logs",
		"internal-json",
		"bar",
		"bar-with-logs",
	}
	for _, format := range formats {
		if got := SetLogFormat(ctx, format); got != NixOk {
			t.Fatalf("SetLogFormat(%q) = %v, want %v: %s", format, got, NixOk, errMsgString(t, ctx))
		}
	}
}

func TestNixMainSetLogFormatInvalidSetsContextError(t *testing.T) {
	ctx := newTestMainContext(t)

	got := SetLogFormat(ctx, "go-bindings-test-invalid-log-format")
	if got == NixOk {
		t.Fatalf("SetLogFormat(invalid) = %v, want non-OK", got)
	}
	if errCode := ErrCode(ctx); errCode == NixOk {
		t.Fatalf("ErrCode after invalid SetLogFormat = %v, want non-OK", errCode)
	}
	if msg := errMsgString(t, ctx); !strings.Contains(msg, "invalid value") {
		t.Fatalf("ErrMsg after invalid SetLogFormat = %q, want invalid value message", msg)
	}
}

func TestNixMainInitPluginsDefaultEmptyConfig(t *testing.T) {
	ctx := newTestMainContext(t)

	if got := InitPlugins(ctx); got != NixOk {
		t.Fatalf("InitPlugins = %v, want %v: %s", got, NixOk, errMsgString(t, ctx))
	}
}
