package api

import (
	"testing"
	"time"
)

func TestRFC6238SHA1Vectors(t *testing.T) {
	secret := []byte("12345678901234567890")
	for _, v := range []struct {
		at   int64
		code string
	}{{59, "94287082"}, {1111111109, "07081804"}, {1111111111, "14050471"}, {1234567890, "89005924"}, {2000000000, "69279037"}, {20000000000, "65353130"}} {
		if got := totpCode(secret, v.at/30, 8); got != v.code {
			t.Fatalf("time %d: %s != %s", v.at, got, v.code)
		}
	}
	now := time.Unix(1234567890, 0)
	n := now.Unix() / 30
	code := totpCode(secret, n, 6)
	if _, ok := matchTOTP(secret, code, now, n-1); !ok {
		t.Fatal("valid code rejected")
	}
	if _, ok := matchTOTP(secret, code, now, n); ok {
		t.Fatal("replay accepted")
	}
	if _, ok := matchTOTP(secret, totpCode(secret, n-2, 6), now, -1); ok {
		t.Fatal("expired code accepted")
	}
	for _, invalid := range []string{"", "12345", "1234567", "abcdef"} {
		if _, ok := matchTOTP(secret, invalid, now, -1); ok {
			t.Fatal("invalid code accepted")
		}
	}
}
