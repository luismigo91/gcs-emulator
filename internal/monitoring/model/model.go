package model

import "time"

type TimeSeries struct {
	Metric     *Metric            `json:"metric"`
	Resource   *MonitoredResource `json:"resource"`
	MetricKind string             `json:"metricKind,omitempty"`
	ValueType  string             `json:"valueType,omitempty"`
	Points     []*Point           `json:"points"`
}

type Metric struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type MonitoredResource struct {
	Type   string            `json:"type"`
	Labels map[string]string `json:"labels"`
}

type Point struct {
	Interval *TimeInterval    `json:"interval"`
	Value    *TypedValue      `json:"value"`
}

type TimeInterval struct {
	StartTime time.Time `json:"startTime,omitempty"`
	EndTime   time.Time `json:"endTime"`
}

type TypedValue struct {
	BoolValue         *bool    `json:"boolValue,omitempty"`
	Int64Value        *int64   `json:"int64Value,omitempty"`
	DoubleValue       *float64 `json:"doubleValue,omitempty"`
	StringValue       *string  `json:"stringValue,omitempty"`
	DistributionValue *Distribution `json:"distributionValue,omitempty"`
}

type Distribution struct {
	Count        int64     `json:"count"`
	Mean         float64   `json:"mean"`
	BucketCounts []int64   `json:"bucketCounts"`
	BucketOptions *BucketOptions `json:"bucketOptions,omitempty"`
}

type BucketOptions struct {
	ExplicitBuckets *Explicit `json:"explicitBuckets,omitempty"`
}

type Explicit struct {
	Bounds []float64 `json:"bounds"`
}

type CreateTimeSeriesRequest struct {
	TimeSeries []*TimeSeries `json:"timeSeries"`
}

type ListTimeSeriesRequest struct {
	Name     string `json:"name"`
	Filter   string `json:"filter"`
	Interval *TimeInterval `json:"interval"`
}
