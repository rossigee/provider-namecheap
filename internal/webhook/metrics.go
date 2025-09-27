package webhook

import (
	"sync"
	"time"
)

// Metrics provides observability for webhook operations
type Metrics struct {
	mu                sync.RWMutex
	RequestsTotal     *Counter
	RequestsErrors    *Counter
	ProcessingErrors  *Counter
	EventsProcessed   *Counter
	RequestDuration   *Histogram
	lastReset         time.Time
}

// Counter represents a simple counter metric
type Counter struct {
	mu    sync.RWMutex
	value int64
}

// Inc increments the counter by 1
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Add increments the counter by the given value
func (c *Counter) Add(v int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += v
}

// Value returns the current counter value
func (c *Counter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Histogram tracks request duration metrics
type Histogram struct {
	mu      sync.RWMutex
	samples []float64
	sum     float64
	count   int64
}

// Observe records a new sample
func (h *Histogram) Observe(v float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.samples = append(h.samples, v)
	h.sum += v
	h.count++

	// Keep only recent samples to prevent memory growth
	if len(h.samples) > 1000 {
		h.samples = h.samples[len(h.samples)-500:]
	}
}

// Average returns the average duration
func (h *Histogram) Average() float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.count == 0 {
		return 0
	}
	return h.sum / float64(h.count)
}

// Count returns the number of observations
func (h *Histogram) Count() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		RequestsTotal:    &Counter{},
		RequestsErrors:   &Counter{},
		ProcessingErrors: &Counter{},
		EventsProcessed:  &Counter{},
		RequestDuration:  &Histogram{},
		lastReset:        time.Now(),
	}
}

// GetAll returns all metrics as a map for JSON serialization
func (m *Metrics) GetAll() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"requests_total":        m.RequestsTotal.Value(),
		"requests_errors":       m.RequestsErrors.Value(),
		"processing_errors":     m.ProcessingErrors.Value(),
		"events_processed":      m.EventsProcessed.Value(),
		"request_duration_avg":  m.RequestDuration.Average(),
		"request_count":         m.RequestDuration.Count(),
		"uptime_seconds":        time.Since(m.lastReset).Seconds(),
		"last_reset":            m.lastReset.Format(time.RFC3339),
	}
}

// Reset resets all metrics counters
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestsTotal = &Counter{}
	m.RequestsErrors = &Counter{}
	m.ProcessingErrors = &Counter{}
	m.EventsProcessed = &Counter{}
	m.RequestDuration = &Histogram{}
	m.lastReset = time.Now()
}