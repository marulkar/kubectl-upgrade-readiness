package kubelet

import (
	"testing"

	"github.com/blang/semver/v4"
)

func TestKubeletVersionSkew(t *testing.T) {
	target, _ := semver.Parse("1.31.0")

	tests := []struct {
		name     string
		kubelet  string
		expected bool // true if within skew
	}{
		{"EqualVersion", "v1.31.0", true},
		{"WithinSkewN-1", "v1.30.1", true},
		{"WithinSkewN-4", "v1.27.5", true},
		{"TooOldSkew", "v1.26.9", false},
		{"TooNew", "v1.32.0", false},
		{"DifferentMajor", "v2.30.0", false},
		{"MalformedVersion", "not-a-version", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v, err := semver.ParseTolerant(tc.kubelet)
			ok := err == nil && v.Major == target.Major && v.Minor <= target.Minor && target.Minor-v.Minor <= 4
			if ok != tc.expected {
				t.Errorf("Version %s => expected %v, got %v", tc.kubelet, tc.expected, ok)
			}
		})
	}
}
