package gamma_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johannestaas/gamma"
)

func testClient() *http.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	return server.Client()
}

func TestClientSetup(t *testing.T) {
	opts := []gamma.Option{
		gamma.WithRootURL("http://example.org"),
		gamma.WithModel("foobar"),
		gamma.WithHTTPClient(testClient()),
	}
	c := gamma.NewGammaClient(opts...)
	if c.RootURL != "http://example.org" {
		t.Fatalf("root url was: %+v", c.RootURL)
	}
	if c.Model != "foobar" {
		t.Fatalf("client model was: %+v", c.Model)
	}
}
