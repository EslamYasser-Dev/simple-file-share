package metrics

import (
	"strings"
	"sync"
	"testing"
)

func render(t *testing.T, r *Registry) string {
	t.Helper()
	var sb strings.Builder
	if err := r.WriteText(&sb); err != nil {
		t.Fatalf("WriteText: %v", err)
	}
	return sb.String()
}

func TestCounterExposition(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("fs_requests_total", "Total requests.", "method", "route")
	c.WithLabelValues("GET", "/api/files").Inc()
	c.WithLabelValues("GET", "/api/files").Add(2)
	c.WithLabelValues("POST", "/api/files").Inc()

	out := render(t, r)
	for _, want := range []string{
		"# HELP fs_requests_total Total requests.",
		"# TYPE fs_requests_total counter",
		`fs_requests_total{method="GET",route="/api/files"} 3`,
		`fs_requests_total{method="POST",route="/api/files"} 1`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestCounterNegativeIsNoOp(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("fs_x_total", "x").WithLabelValues()
	c.Add(5)
	c.Add(-3)
	out := render(t, r)
	if !strings.Contains(out, "fs_x_total 5") {
		t.Fatalf("negative add should be ignored:\n%s", out)
	}
}

func TestGaugeSetAddDec(t *testing.T) {
	r := NewRegistry()
	g := r.NewGauge("fs_active", "Active streams").WithLabelValues()
	g.Inc()
	g.Inc()
	g.Dec()
	g.Set(7)
	out := render(t, r)
	if !strings.Contains(out, "fs_active 7") {
		t.Fatalf("gauge = \n%s", out)
	}
}

func TestHistogramBucketsCumulative(t *testing.T) {
	r := NewRegistry()
	h := r.NewHistogram("fs_latency_seconds", "Latency.", []float64{0.1, 0.5, 1}, "route")
	for _, v := range []float64{0.05, 0.2, 0.7, 2} {
		h.WithLabelValues("/a").Observe(v)
	}
	out := render(t, r)
	for _, want := range []string{
		"# TYPE fs_latency_seconds histogram",
		`fs_latency_seconds_bucket{route="/a",le="0.1"} 1`,
		`fs_latency_seconds_bucket{route="/a",le="0.5"} 2`,
		`fs_latency_seconds_bucket{route="/a",le="1"} 3`,
		`fs_latency_seconds_bucket{route="/a",le="+Inf"} 4`,
		`fs_latency_seconds_count{route="/a"} 4`,
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("histogram output missing %q:\n%s", want, out)
		}
	}
}

func TestLabelCountMismatchIsNoOp(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("fs_y_total", "y", "a", "b")
	c.WithLabelValues("only-one").Inc() // wrong arity: must not panic or record
	c.WithLabelValues("a", "b").Inc()
	out := render(t, r)
	if strings.Contains(out, `fs_y_total{a="only-one"}`) {
		t.Fatalf("mismatched labels recorded:\n%s", out)
	}
	if !strings.Contains(out, `fs_y_total{a="a",b="b"} 1`) {
		t.Fatalf("valid series missing:\n%s", out)
	}
}

func TestInvalidNamesAreDroppedNotFatal(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("bad-name", "no dashes allowed")
	r.NewCounter("good_total", "fine", "1badlabel")
	r.NewGauge("also-bad", "x")
	if out := render(t, r); strings.Contains(out, "bad-name") || strings.Contains(out, "also-bad") {
		t.Fatalf("invalid metrics must not render:\n%s", out)
	}
}

func TestDuplicateRegistrationKeepsFirst(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("fs_dup_total", "first").WithLabelValues().Inc()
	r.NewCounter("fs_dup_total", "second").WithLabelValues().Inc()
	out := render(t, r)
	if strings.Count(out, "# TYPE fs_dup_total") != 1 {
		t.Fatalf("duplicate family rendered twice:\n%s", out)
	}
	if !strings.Contains(out, "# HELP fs_dup_total first") {
		t.Fatalf("second registration replaced the first:\n%s", out)
	}
}

func TestLabelValueEscaping(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("fs_esc_total", "esc", "path").WithLabelValues(`a"b\c`).Inc()
	out := render(t, r)
	if !strings.Contains(out, `path="a\"b\\c"`) {
		t.Fatalf("label not escaped:\n%s", out)
	}
}

func TestHelpEscaping(t *testing.T) {
	r := NewRegistry()
	r.NewCounter("fs_help_total", "line\nbreak")
	out := render(t, r)
	if !strings.Contains(out, `# HELP fs_help_total line\nbreak`) {
		t.Fatalf("help not escaped:\n%s", out)
	}
}

func TestDeterministicOrdering(t *testing.T) {
	for i := 0; i < 20; i++ {
		r := NewRegistry()
		g := r.NewGauge("fs_order", "o", "k")
		for _, v := range []string{"zeta", "alpha", "mid", "beta"} {
			g.WithLabelValues(v).Set(1)
		}
		out := render(t, r)
		ia, iz := strings.Index(out, `k="alpha"`), strings.Index(out, `k="zeta"`)
		if ia < 0 || iz < 0 || ia > iz {
			t.Fatalf("series not sorted:\n%s", out)
		}
	}
}

func TestConcurrentUseAndScrape(t *testing.T) {
	r := NewRegistry()
	c := r.NewCounter("fs_conc_total", "c", "id")
	h := r.NewHistogram("fs_conc_seconds", "h", []float64{0.1, 1}, "id")
	g := r.NewGauge("fs_conc_g", "g").WithLabelValues()

	var wg sync.WaitGroup
	for w := 0; w < 8; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			id := string(rune('a' + w))
			for i := 0; i < 100; i++ {
				c.WithLabelValues(id).Inc()
				h.WithLabelValues(id).Observe(0.05)
				g.Inc()
				if i%10 == 0 {
					var sb strings.Builder
					_ = r.WriteText(&sb)
				}
			}
		}(w)
	}
	wg.Wait()

	out := render(t, r)
	for w := 0; w < 8; w++ {
		if want := `fs_conc_total{id="` + string(rune('a'+w)) + `"} 100`; !strings.Contains(out, want) {
			t.Fatalf("lost counts, missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "fs_conc_g 800") {
		t.Fatalf("gauge lost updates:\n%s", out)
	}
}
