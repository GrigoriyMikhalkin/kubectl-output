package plugin

import "testing"

func TestSplitResourceName(t *testing.T) {
	typ, version, group := splitResourceName("pods")
	if typ != "pods" || version != "" || group != "" {
		t.Fatalf("expected: pods, got: %s.%s.%s", typ, version, group)
	}

	typ, version, group = splitResourceName("pods.v1")
	if typ != "pods" || version != "v1" || group != "" {
		t.Fatalf("expected: pods.v1, got: %s.%s.%s", typ, version, group)
	}

	typ, version, group = splitResourceName("http.routes.test.io")
	if typ != "http" || version != "" || group != "routes.test.io" {
		t.Fatalf("expected: http.routes.test.io, got: %s.%s.%s", typ, version, group)
	}

	typ, version, group = splitResourceName("http.v1.routes.test.io")
	if typ != "http" || version != "v1" || group != "routes.test.io" {
		t.Fatalf("expected: http.v1.routes.test.io, got: %s.%s.%s", typ, version, group)
	}
}
