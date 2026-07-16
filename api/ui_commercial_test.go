package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/migrations"
)

// TestCommercialUIMonochromeShell exercises the real Gin router, temp SQLite DB,
// and shipped CSS so the commercial monochrome shell cannot regress silently.
func TestCommercialUIMonochromeShell(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ui-commercial.db")
	if err := database.Init(path, "silent"); err != nil {
		t.Fatalf("init database: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
	if err := migrations.Run(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	router, err := NewRouter()
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	// Core product routes must render the app shell.
	routes := []string{"/", "/targets", "/scenarios", "/runs"}
	for _, route := range routes {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, route, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d body=%s", route, rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		for _, needle := range []string{
			`class="app-frame"`,
			`class="sidebar"`,
			`class="topbar"`,
			`class="shell"`,
			`class="site-footer"`,
			`/static/app.css`,
			`/static/pages.css`,
		} {
			if !strings.Contains(body, needle) {
				t.Fatalf("%s: missing shell marker %q", route, needle)
			}
		}
		// Metric labels should not dump long glossary walls into the page.
		if strings.Contains(body, "指标说明") || strings.Contains(body, "glossary-grid") {
			t.Fatalf("%s: unexpected dense glossary markup", route)
		}
	}

	// Shipped design tokens must stay monochrome commercial (black primary, light borders).
	cssRec := httptest.NewRecorder()
	router.ServeHTTP(cssRec, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	if cssRec.Code != http.StatusOK {
		t.Fatalf("app.css: expected 200, got %d", cssRec.Code)
	}
	css := cssRec.Body.String()
	required := []string{
		"--fg: #000000",
		"--accent: #000000",
		"--border: #eaeaea",
		"--bg: #fafafa",
		"font-size: 14px",
		"overflow: hidden",
		".app-frame",
		".shell",
		"overflow-y: auto",
	}
	for _, token := range required {
		if !strings.Contains(css, token) {
			t.Fatalf("app.css missing required commercial token/rule %q", token)
		}
	}

	// Brand-primary blue gradients must not reappear on core shell styles.
	blueBrand := regexp.MustCompile(`#2563eb|#3b82f6|#1d4ed8|linear-gradient\([^)]*#3b82f6`)
	if blueBrand.MatchString(css) {
		t.Fatalf("app.css contains brand-primary blue accents: %s", blueBrand.FindString(css))
	}

	// Report CSS should also stay monochrome for chart accents.
	reportRec := httptest.NewRecorder()
	router.ServeHTTP(reportRec, httptest.NewRequest(http.MethodGet, "/static/report.css", nil))
	if reportRec.Code != http.StatusOK {
		t.Fatalf("report.css: expected 200, got %d", reportRec.Code)
	}
	if blueBrand.MatchString(reportRec.Body.String()) {
		t.Fatalf("report.css contains brand-primary blue accents")
	}

	// Source file on disk must match served monochrome contract.
	root := findModuleRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "web", "static", "app.css"))
	if err != nil {
		t.Fatalf("read app.css: %v", err)
	}
	if !strings.Contains(string(raw), "--accent: #000000") {
		t.Fatalf("source app.css lost monochrome accent token")
	}

	// Merged run detail must keep short labels with field-help (live log + report).
	runHTML, err := os.ReadFile(filepath.Join(root, "web", "templates", "run_detail.html"))
	if err != nil {
		t.Fatalf("read run_detail.html: %v", err)
	}
	runSrc := string(runHTML)
	if !strings.Contains(runSrc, `data-help=`) || !strings.Contains(runSrc, `class="field-help"`) {
		t.Fatal("run_detail.html must use field-help tooltips for metric explanations")
	}
	if !strings.Contains(runSrc, "run-merged") || !strings.Contains(runSrc, "请求日志") {
		t.Fatal("run_detail.html must merge live log and performance report")
	}
	if strings.Contains(runSrc, "glossary-grid") || strings.Contains(runSrc, "指标说明") {
		t.Fatal("run_detail.html must not reintroduce dense glossary markup")
	}
	for _, label := range []string{"E2E", "TTFT", "TPOT", "Queue", "RPS"} {
		if !strings.Contains(runSrc, label) {
			t.Fatalf("run_detail.html missing short metric label %q", label)
		}
	}

	// /reports/:id should redirect to the merged /runs/:id page.
	redirectRec := httptest.NewRecorder()
	router.ServeHTTP(redirectRec, httptest.NewRequest(http.MethodGet, "/reports/1", nil))
	if redirectRec.Code != http.StatusFound {
		t.Fatalf("reports redirect: expected 302, got %d", redirectRec.Code)
	}
	if loc := redirectRec.Header().Get("Location"); loc != "/runs/1" {
		t.Fatalf("reports redirect location: got %q want /runs/1", loc)
	}
}

func findModuleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getcwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from working directory")
		}
		dir = parent
	}
}
