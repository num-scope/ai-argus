package utils

import "testing"

func TestNormalizePageAppliesDefaultsAndLimit(t *testing.T) {
	page, pageSize := NormalizePage(0, 0)
	if page != 1 || pageSize != 20 {
		t.Fatalf("unexpected defaults: page=%d pageSize=%d", page, pageSize)
	}
	page, pageSize = NormalizePage(3, 1000)
	if page != 3 || pageSize != 100 {
		t.Fatalf("unexpected limit: page=%d pageSize=%d", page, pageSize)
	}
}
