package api

import (
	"net/http"
	"testing"

	"github.com/KilianLhy/mini-claude-go/internal/shared"
)

func TestRateLimiterAllowsBurstThenBlocks(t *testing.T) {
	rl := newRateLimiter(0, 3)

	for i := 0; i < 3; i++ {
		if !rl.allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed within the burst", i+1)
		}
	}
	if rl.allow("1.2.3.4") {
		t.Fatal("request beyond the burst should be blocked")
	}
	if !rl.allow("5.6.7.8") {
		t.Fatal("a different client should not be affected")
	}
}

func TestLoginIsRateLimited(t *testing.T) {
	r := newTestRouter()
	body := shared.LoginRequest{Email: "nobody@example.com", Password: "whatever12"}

	var got429 bool
	for i := 0; i < 10; i++ {
		rec := do(t, r, http.MethodPost, shared.RouteLogin, "", body)
		if rec.Code == http.StatusTooManyRequests {
			got429 = true
			break
		}
	}
	if !got429 {
		t.Fatal("expected a 429 Too Many Requests after repeated login attempts")
	}
}
