package antigravity

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dvcrn/antigravity-oauth-proxy/internal/credentials"
)

func TestRetrieveUserQuotaSummary(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		_, _ = w.Write([]byte(`{"groups":[
			{"displayName":"Gemini Models","buckets":[
				{"bucketId":"gemini-weekly","window":"weekly","resetTime":"2026-09-29T09:12:52Z","remainingFraction":0.8564},
				{"bucketId":"gemini-5h","window":"5h","resetTime":"2026-09-22T14:12:52Z"}]},
			{"displayName":"Claude and GPT models","buckets":[
				{"bucketId":"3p-weekly","window":"weekly","remainingFraction":0},
				{"bucketId":"3p-5h","window":"5h","remainingFraction":1}]}]}`))
	}))
	defer srv.Close()

	orig := Endpoints
	Endpoints = []string{srv.URL}
	defer func() { Endpoints = orig }()

	client := NewClient(&refreshTestProvider{creds: credentials.OAuthCredentials{
		AccessToken: "token",
		ExpiryDate:  time.Now().Add(time.Hour).UnixMilli(),
	}})
	summary, err := client.RetrieveUserQuotaSummary(context.Background(), "proj-1")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1internal:retrieveUserQuotaSummary" || gotBody != `{"project":"proj-1"}` {
		t.Fatalf("request = %s %s", gotPath, gotBody)
	}
	if len(summary.Groups) != 2 || len(summary.Groups[0].Buckets) != 2 {
		t.Fatalf("groups = %+v", summary.Groups)
	}
	weekly := summary.Groups[0].Buckets[0]
	if weekly.Window != "weekly" || weekly.RemainingFraction == nil || *weekly.RemainingFraction != 0.8564 {
		t.Fatalf("gemini weekly = %+v", weekly)
	}
	if summary.Groups[0].Buckets[1].RemainingFraction != nil {
		t.Fatal("missing remainingFraction should stay nil")
	}
	if r := summary.Groups[1].Buckets[0].RemainingFraction; r == nil || *r != 0 {
		t.Fatal("explicit 0 remainingFraction should be kept")
	}
}
