package ports

import "time"

// AnalyticsEvent is one PII-free usage record persisted for admin reporting.
// Path is the virtual path the client used; no credentials or IPs are stored.
type AnalyticsEvent struct {
	Type  string    `json:"type"`
	Path  string    `json:"path,omitempty"`
	User  string    `json:"user,omitempty"`
	Bytes int64     `json:"bytes,omitempty"`
	At    time.Time `json:"at"`
}

// AnalyticsOverview is the rollup for a time window.
type AnalyticsOverview struct {
	TotalEvents   int64            `json:"totalEvents"`
	Uploads       int64            `json:"uploads"`
	Downloads     int64            `json:"downloads"`
	Deletes       int64            `json:"deletes"`
	Shares        int64            `json:"shares"`
	Logins        int64            `json:"logins"`
	ActiveUsers   int64            `json:"activeUsers"`
	BytesUploaded int64            `json:"bytesUploaded"`
	ByType        map[string]int64 `json:"byType"`
	WindowStart   time.Time        `json:"windowStart"`
	WindowEnd     time.Time        `json:"windowEnd"`
}

// AnalyticsBucket is one timeline slice (day or hour).
type AnalyticsBucket struct {
	Start time.Time `json:"start"`
	Count int64     `json:"count"`
	Bytes int64     `json:"bytes"`
}

// AnalyticsTopFile is a frequently touched path in the window.
type AnalyticsTopFile struct {
	Path  string `json:"path"`
	Count int64  `json:"count"`
	Bytes int64  `json:"bytes"`
}

// AnalyticsStore appends usage events and aggregates them on read.
type AnalyticsStore interface {
	Record(e AnalyticsEvent) error
	Overview(since, until time.Time) (*AnalyticsOverview, error)
	Timeline(since, until time.Time) ([]AnalyticsBucket, error)
	TopFiles(since, until time.Time, limit int) ([]AnalyticsTopFile, error)
}
