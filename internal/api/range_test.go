package api_test

import (
	"strconv"
	"strings"
	"testing"
)

func parseRangeTest(rangeHeaderValue string, contentLength int64) (start int64, end int64, satisfiable bool) {
	parts := strings.SplitN(rangeHeaderValue, "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, contentLength - 1, true
	}

	rangeSpec := parts[1]
	if len(rangeSpec) == 0 {
		return 0, contentLength - 1, true
	}

	if rangeSpec[0] == '-' {
		offsetFromEnd, err := strconv.ParseInt(rangeSpec, 10, 64)
		if err != nil {
			return 0, contentLength - 1, true
		}
		start = contentLength + offsetFromEnd
		if start < 0 {
			start = 0
		}
		return start, contentLength - 1, true
	}

	rangeParts := strings.SplitN(rangeSpec, "-", 2)
	if len(rangeParts) != 2 {
		return 0, contentLength - 1, true
	}

	var s int64
	s, err := strconv.ParseInt(rangeParts[0], 10, 64)
	if err != nil {
		return 0, contentLength - 1, true
	}

	if rangeParts[1] == "" {
		return s, contentLength - 1, s < contentLength
	}

	var e int64
	e, err = strconv.ParseInt(rangeParts[1], 10, 64)
	if err != nil {
		return 0, contentLength - 1, true
	}

	if e >= contentLength {
		e = contentLength - 1
	}

	if s >= contentLength {
		return s, e, false
	}

	if e < s {
		return 0, contentLength - 1, true
	}

	return s, e, true
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		name        string
		rangeHeader string
		contentLen  int64
		wantStart   int64
		wantEnd     int64
		wantSatisf  bool
	}{
		{
			name:        "simple range",
			rangeHeader: "bytes=0-99",
			contentLen:  1000,
			wantStart:   0,
			wantEnd:     99,
			wantSatisf:  true,
		},
		{
			name:        "range to end",
			rangeHeader: "bytes=500-",
			contentLen:  1000,
			wantStart:   500,
			wantEnd:     999,
			wantSatisf:  true,
		},
		{
			name:        "suffix range",
			rangeHeader: "bytes=-100",
			contentLen:  1000,
			wantStart:   900,
			wantEnd:     999,
			wantSatisf:  true,
		},
		{
			name:        "range beyond content",
			rangeHeader: "bytes=2000-3000",
			contentLen:  1000,
			wantStart:   2000,
			wantEnd:     999,
			wantSatisf:  false,
		},
		{
			name:        "invalid range format",
			rangeHeader: "invalid",
			contentLen:  1000,
			wantStart:   0,
			wantEnd:     999,
			wantSatisf:  true,
		},
		{
			name:        "wrong unit",
			rangeHeader: "bits=0-100",
			contentLen:  1000,
			wantStart:   0,
			wantEnd:     999,
			wantSatisf:  true,
		},
		{
			name:        "end before start",
			rangeHeader: "bytes=100-50",
			contentLen:  1000,
			wantStart:   0,
			wantEnd:     999,
			wantSatisf:  true,
		},
		{
			name:        "end clamped to content length",
			rangeHeader: "bytes=900-2000",
			contentLen:  1000,
			wantStart:   900,
			wantEnd:     999,
			wantSatisf:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, satisfiable := parseRangeTest(tt.rangeHeader, tt.contentLen)
			if start != tt.wantStart {
				t.Errorf("start = %d, want %d", start, tt.wantStart)
			}
			if end != tt.wantEnd {
				t.Errorf("end = %d, want %d", end, tt.wantEnd)
			}
			if satisfiable != tt.wantSatisf {
				t.Errorf("satisfiable = %v, want %v", satisfiable, tt.wantSatisf)
			}
		})
	}
}
