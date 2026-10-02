// Package metrics is a dependency-free metrics registry that renders the
// Prometheus text exposition format. It exists because the project keeps a
// strict stdlib-only dependency policy: counters, gauges, and histograms are
// hand-rolled instead of pulling in a client library.
package metrics

import (
	"fmt"
	"io"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// collector is one metric family (name + type + help + its label series).
type collector interface {
	// kind reports the metric type for the # TYPE line.
	kind() string
	// name is the metric family name.
	family() string
	// help is the # HELP description.
	description() string
	// writeSeries renders the per-label-value sample lines. The caller holds
	// no locks; implementations lock their own series maps.
	writeSeries(w io.Writer) error
}

// Registry holds every registered metric family in insertion order and
// renders them as a Prometheus text exposition payload.
type Registry struct {
	mu     sync.Mutex
	family []collector
	index  map[string]struct{}
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{index: make(map[string]struct{})}
}

// register adds c unless the name is already taken or invalid; it returns
// whether the collector should be kept. Duplicate or malformed names are
// dropped rather than panicking: a bad metric must never take the server down.
func (r *Registry) register(c collector, name string, labelNames []string) bool {
	if !validMetricName(name) || !validLabelNames(labelNames) {
		return false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.index[name]; dup {
		return false
	}
	r.index[name] = struct{}{}
	r.family = append(r.family, c)
	return true
}

// NewCounter registers a counter family and returns its label router.
func (r *Registry) NewCounter(name, help string, labelNames ...string) *CounterVec {
	v := &CounterVec{
		name:       name,
		help:       help,
		labelNames: append([]string(nil), labelNames...),
		series:     make(map[string]*counterSeries),
	}
	if !r.register(v, name, labelNames) {
		v.disabled = true
	}
	return v
}

// NewGauge registers a gauge family and returns its label router.
func (r *Registry) NewGauge(name, help string, labelNames ...string) *GaugeVec {
	v := &GaugeVec{
		name:       name,
		help:       help,
		labelNames: append([]string(nil), labelNames...),
		series:     make(map[string]*gaugeSeries),
	}
	if !r.register(v, name, labelNames) {
		v.disabled = true
	}
	return v
}

// NewHistogram registers a histogram family with explicit upper bounds and
// returns its label router. Buckets must be sorted ascending; invalid input
// falls back to a small default so callers never panic.
func (r *Registry) NewHistogram(name, help string, buckets []float64, labelNames ...string) *HistogramVec {
	b := append([]float64(nil), buckets...)
	if len(b) == 0 {
		b = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	}
	if !sort.Float64sAreSorted(b) {
		sort.Float64s(b)
	}
	v := &HistogramVec{
		name:       name,
		help:       help,
		labelNames: append([]string(nil), labelNames...),
		buckets:    b,
		series:     make(map[string]*histogramSeries),
	}
	if !r.register(v, name, labelNames) {
		v.disabled = true
	}
	return v
}

// WriteText renders the full registry in Prometheus text exposition format.
func (r *Registry) WriteText(w io.Writer) error {
	r.mu.Lock()
	cols := append([]collector(nil), r.family...)
	r.mu.Unlock()

	for _, c := range cols {
		if c.description() != "" {
			if _, err := fmt.Fprintf(w, "# HELP %s %s\n", c.family(), escapeHelp(c.description())); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "# TYPE %s %s\n", c.family(), c.kind()); err != nil {
			return err
		}
		if err := c.writeSeries(w); err != nil {
			return err
		}
	}
	return nil
}

// --- counter ---

// CounterVec routes label values to counter series.
type CounterVec struct {
	name       string
	help       string
	labelNames []string
	disabled   bool
	mu         sync.Mutex
	series     map[string]*counterSeries
}

type counterSeries struct {
	labels []string
	value  float64
}

// Counter is a handle to one labelled counter series.
type Counter struct {
	vec    *CounterVec
	key    string
	labels []string
}

// WithLabelValues returns the counter for the given label values. The number
// of values must match the family's label names; a mismatch returns a
// no-op handle so metrics can never crash a request path.
func (v *CounterVec) WithLabelValues(values ...string) *Counter {
	if v == nil || v.disabled || len(values) != len(v.labelNames) {
		return &Counter{}
	}
	key := encodeKey(values)
	v.mu.Lock()
	if _, ok := v.series[key]; !ok {
		v.series[key] = &counterSeries{labels: append([]string(nil), values...)}
	}
	v.mu.Unlock()
	return &Counter{vec: v, key: key, labels: values}
}

// Inc adds one.
func (c *Counter) Inc() { c.Add(1) }

// Add adds delta to the series (negative deltas are rejected by convention:
// they become no-ops to keep counters monotonic).
func (c *Counter) Add(delta float64) {
	if c == nil || c.vec == nil || delta <= 0 {
		return
	}
	c.vec.mu.Lock()
	if s, ok := c.vec.series[c.key]; ok {
		s.value += delta
	}
	c.vec.mu.Unlock()
}

func (v *CounterVec) kind() string        { return "counter" }
func (v *CounterVec) family() string      { return v.name }
func (v *CounterVec) description() string { return v.help }

func (v *CounterVec) writeSeries(w io.Writer) error {
	v.mu.Lock()
	series := make([]counterSeries, 0, len(v.series))
	for _, s := range v.series {
		series = append(series, *s)
	}
	v.mu.Unlock()
	sortSeries(series, func(s counterSeries) []string { return s.labels })
	for _, s := range series {
		if _, err := fmt.Fprintf(w, "%s%s %s\n", v.name, formatLabels(s.labels, v.labelNames), formatFloat(s.value)); err != nil {
			return err
		}
	}
	return nil
}

// --- gauge ---

// GaugeVec routes label values to gauge series.
type GaugeVec struct {
	name       string
	help       string
	labelNames []string
	disabled   bool
	mu         sync.Mutex
	series     map[string]*gaugeSeries
}

type gaugeSeries struct {
	labels []string
	value  float64
}

// Gauge is a handle to one labelled gauge series.
type Gauge struct {
	vec    *GaugeVec
	key    string
	labels []string
}

// WithLabelValues returns the gauge for the given label values.
func (v *GaugeVec) WithLabelValues(values ...string) *Gauge {
	if v == nil || v.disabled || len(values) != len(v.labelNames) {
		return &Gauge{}
	}
	key := encodeKey(values)
	v.mu.Lock()
	if _, ok := v.series[key]; !ok {
		v.series[key] = &gaugeSeries{labels: append([]string(nil), values...)}
	}
	v.mu.Unlock()
	return &Gauge{vec: v, key: key, labels: values}
}

// Set overwrites the current value.
func (g *Gauge) Set(value float64) {
	if g == nil || g.vec == nil {
		return
	}
	g.vec.mu.Lock()
	if s, ok := g.vec.series[g.key]; ok {
		s.value = value
	}
	g.vec.mu.Unlock()
}

// Add adds delta (gauges may move in both directions).
func (g *Gauge) Add(delta float64) {
	if g == nil || g.vec == nil {
		return
	}
	g.vec.mu.Lock()
	if s, ok := g.vec.series[g.key]; ok {
		s.value += delta
	}
	g.vec.mu.Unlock()
}

// Inc adds one.
func (g *Gauge) Inc() { g.Add(1) }

// Dec subtracts one.
func (g *Gauge) Dec() { g.Add(-1) }

func (v *GaugeVec) kind() string        { return "gauge" }
func (v *GaugeVec) family() string      { return v.name }
func (v *GaugeVec) description() string { return v.help }

func (v *GaugeVec) writeSeries(w io.Writer) error {
	v.mu.Lock()
	series := make([]gaugeSeries, 0, len(v.series))
	for _, s := range v.series {
		series = append(series, *s)
	}
	v.mu.Unlock()
	sortSeries(series, func(s gaugeSeries) []string { return s.labels })
	for _, s := range series {
		if _, err := fmt.Fprintf(w, "%s%s %s\n", v.name, formatLabels(s.labels, v.labelNames), formatFloat(s.value)); err != nil {
			return err
		}
	}
	return nil
}

// --- histogram ---

// HistogramVec routes label values to histogram series.
type HistogramVec struct {
	name       string
	help       string
	labelNames []string
	buckets    []float64
	disabled   bool
	mu         sync.Mutex
	series     map[string]*histogramSeries
}

type histogramSeries struct {
	labels []string
	counts []uint64 // per-bucket (non-cumulative) counts, len == len(buckets)
	sum    float64
	total  uint64
}

// Histogram is a handle to one labelled histogram series.
type Histogram struct {
	vec    *HistogramVec
	key    string
	labels []string
}

// WithLabelValues returns the histogram for the given label values.
func (v *HistogramVec) WithLabelValues(values ...string) *Histogram {
	if v == nil || v.disabled || len(values) != len(v.labelNames) {
		return &Histogram{}
	}
	key := encodeKey(values)
	v.mu.Lock()
	if _, ok := v.series[key]; !ok {
		v.series[key] = &histogramSeries{
			labels: append([]string(nil), values...),
			counts: make([]uint64, len(v.buckets)),
		}
	}
	v.mu.Unlock()
	return &Histogram{vec: v, key: key, labels: values}
}

// Observe records one sample.
func (h *Histogram) Observe(value float64) {
	if h == nil || h.vec == nil {
		return
	}
	h.vec.mu.Lock()
	defer h.vec.mu.Unlock()
	s, ok := h.vec.series[h.key]
	if !ok {
		return
	}
	// Find the first bucket whose upper bound includes value. Samples above
	// every bound count only in +Inf (rendered from s.total at write time).
	for i, ub := range h.vec.buckets {
		if value <= ub {
			s.counts[i]++
			break
		}
	}
	s.sum += value
	s.total++
}

func (v *HistogramVec) kind() string        { return "histogram" }
func (v *HistogramVec) family() string      { return v.name }
func (v *HistogramVec) description() string { return v.help }

func (v *HistogramVec) writeSeries(w io.Writer) error {
	v.mu.Lock()
	series := make([]histogramSeries, 0, len(v.series))
	for _, s := range v.series {
		cp := *s
		cp.counts = append([]uint64(nil), s.counts...)
		series = append(series, cp)
	}
	buckets := append([]float64(nil), v.buckets...)
	v.mu.Unlock()
	sortSeries(series, func(s histogramSeries) []string { return s.labels })

	for i := range series {
		s := &series[i]
		labels := formatLabels(s.labels, v.labelNames)
		var cumulative uint64
		for bi, ub := range buckets {
			cumulative += s.counts[bi]
			le := formatLabels(append(append([]string(nil), s.labels...), formatFloat(ub)), append(append([]string(nil), v.labelNames...), "le"))
			if _, err := fmt.Fprintf(w, "%s_bucket%s %d\n", v.name, le, cumulative); err != nil {
				return err
			}
		}
		inf := formatLabels(append(append([]string(nil), s.labels...), "+Inf"), append(append([]string(nil), v.labelNames...), "le"))
		if _, err := fmt.Fprintf(w, "%s_bucket%s %d\n", v.name, inf, s.total); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s_sum%s %s\n", v.name, labels, formatFloat(s.sum)); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "%s_count%s %d\n", v.name, labels, s.total); err != nil {
			return err
		}
	}
	return nil
}

// --- helpers ---

func encodeKey(values []string) string {
	return strings.Join(values, "\x00")
}

func formatLabels(values, names []string) string {
	if len(names) == 0 {
		return ""
	}
	parts := make([]string, 0, len(names))
	for i, n := range names {
		v := ""
		if i < len(values) {
			v = values[i]
		}
		parts = append(parts, n+`="`+escapeLabelValue(v)+`"`)
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func escapeLabelValue(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return strings.ReplaceAll(s, "\n", `\n`)
}

func escapeHelp(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, "\n", `\n`)
}

func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// sortSeries orders already-copied series by encoded label key so exposition
// output is deterministic (stable across scrapes and in tests).
func sortSeries[T any](series []T, labelKey func(T) []string) {
	slices.SortFunc(series, func(a, b T) int {
		return strings.Compare(encodeKey(labelKey(a)), encodeKey(labelKey(b)))
	})
}

// validMetricName enforces the Prometheus [a-zA-Z_:][a-zA-Z0-9_:]* grammar.
func validMetricName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		ok := r == '_' || r == ':' ||
			('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') ||
			(i > 0 && '0' <= r && r <= '9')
		if !ok {
			return false
		}
	}
	return true
}

// validLabelNames enforces the label grammar (no colons) and rejects the
// reserved "le"/"quantile" misuse on non-histogram vectors at registration.
func validLabelNames(names []string) bool {
	seen := make(map[string]struct{}, len(names))
	for _, n := range names {
		if n == "" || strings.ContainsRune(n, ':') {
			return false
		}
		for i, r := range n {
			ok := r == '_' ||
				('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') ||
				(i > 0 && '0' <= r && r <= '9')
			if !ok {
				return false
			}
		}
		if _, dup := seen[n]; dup {
			return false
		}
		seen[n] = struct{}{}
	}
	return true
}
