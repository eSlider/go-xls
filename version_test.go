package xls

import (
	"regexp"
	"testing"
)

// Version must stay in sync with the git tag for each release (semantic versioning: MAJOR.MINOR.PATCH).
func TestVersion(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must not be empty")
	}
	re := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-.]+)?(\+[0-9A-Za-z-.]+)?$`)
	if !re.MatchString(Version) {
		t.Fatalf("Version %q must be valid semantic versioning", Version)
	}
}

func TestVersionIsReleaseTriple(t *testing.T) {
	// This project ships release tags as plain X.Y.Z (no pre-release suffix on the tag itself).
	re := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	if !re.MatchString(Version) {
		t.Fatalf("Version %q must be MAJOR.MINOR.PATCH for release builds", Version)
	}
}
