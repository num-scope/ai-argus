package web

import "testing"

func TestFormatDurationAdaptiveUnits(t *testing.T) {
	cases := []struct {
		seconds float64
		want    string
	}{
		{0, "0 s"},
		{0.25, "250 ms"},
		{3.2, "3.20 s"},
		{45.6, "45.6 s"},
		{90, "1.5 min"},
		{600, "10 min"},
		{7200, "2.0 h"},
		{90000, "1.0 d"},
	}
	for _, tc := range cases {
		got := formatDuration(tc.seconds)
		if got != tc.want {
			t.Fatalf("formatDuration(%v) = %q, want %q", tc.seconds, got, tc.want)
		}
	}
}

func TestFormatLatencyMSAdaptiveUnits(t *testing.T) {
	cases := []struct {
		ms   float64
		want string
	}{
		{0, "-"},
		{12.5, "12.5 ms"},
		{999.4, "999.4 ms"},
		{1000, "1.00 s"},
		{1729.9, "1.73 s"},
		{45_600, "45.6 s"},
		{90_000, "1.5 min"},
		{600_000, "10 min"},
		{7_200_000, "2.0 h"},
	}
	for _, tc := range cases {
		got := formatLatencyMS(tc.ms)
		if got != tc.want {
			t.Fatalf("formatLatencyMS(%v) = %q, want %q", tc.ms, got, tc.want)
		}
	}
}

func TestTemplatesIncludeFormatDuration(t *testing.T) {
	tpl, err := Templates()
	if err != nil {
		t.Fatalf("Templates: %v", err)
	}
	if tpl.Lookup("run_detail.html") == nil {
		t.Fatal("run_detail.html not loaded")
	}
}
