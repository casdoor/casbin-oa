package util

import "testing"

func TestGetOrgStarsStatistics(t *testing.T) {
	org := "casbin"
	t.Logf("Testing organization: %s", org)

	repoCount, totalStars := GetOrgStarsStatistics(org)

	if repoCount == 0 {
		t.Errorf("Expected to find repositories for org %s, but found 0", org)
	}

	if totalStars == 0 {
		t.Errorf("Expected to find stars for org %s repositories, but found 0", org)
	}

	t.Logf("Found %d repositories with total %d stars", repoCount, totalStars)
}
