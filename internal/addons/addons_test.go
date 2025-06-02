package addons

import (
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
)

func TestAddonMatrixCoverage(t *testing.T) {
	tests := []struct {
		k8sVersion string
		addon      string
		imageTag   string
		wantValid  bool
	}{
		{"v1.28", "etcd", "3.5.9", true},
		{"v1.28", "etcd", "3.5.9-0", true}, // suffix should be stripped
		{"v1.29", "coredns", "1.11.0", true},
		{"v1.30", "metrics-server", "0.6.3", false}, // not in matrix
		{"v1.31", "kube-proxy", "1.27.0", true},     // part of N-4 skew
		{"v1.33", "kube-proxy", "1.28.0", false},    // N-5; should be invalid
	}

	for _, tc := range tests {
		version := semver.MustParse(normalize(tc.k8sVersion))
		key := "v" + version.String()[:4]
		allowed, ok := versionedCompatMatrix[key][tc.addon]
		if !ok {
			t.Fatalf("Addon %s not found in matrix for %s", tc.addon, key)
		}

		actual := isTagAllowed(tc.imageTag, allowed)
		assert.Equalf(t, tc.wantValid, actual, "addon=%s, tag=%s, k8s=%s", tc.addon, tc.imageTag, tc.k8sVersion)
	}
}

func normalize(v string) string {
	s := v
	if s[0] == 'v' {
		s = s[1:]
	}
	if parts := len(splitDots(s)); parts == 2 {
		s += ".0"
	}
	return s
}

func splitDots(s string) []string {
	return split(s, '.')
}

func split(s string, sep rune) []string {
	var result []string
	start := 0
	for i, c := range s {
		if c == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	return append(result, s[start:])
}

func isTagAllowed(tag string, allowed []string) bool {
	base := tag
	if parts := strings.SplitN(tag, "-", 2); len(parts) > 1 {
		base = parts[0]
	}
	if strings.Count(base, ".") == 1 {
		base += ".0"
	}
	for _, a := range allowed {
		if a == base {
			return true
		}
	}
	return false
}
