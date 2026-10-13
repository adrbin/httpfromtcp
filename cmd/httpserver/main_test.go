package main

import "testing"

func TestParseByteRange(t *testing.T) {
	tests := []struct {
		value     string
		size      int
		wantStart int
		wantEnd   int
		wantValid bool
	}{
		{value: "bytes=2-5", size: 10, wantStart: 2, wantEnd: 6, wantValid: true},
		{value: "bytes=7-", size: 10, wantStart: 7, wantEnd: 10, wantValid: true},
		{value: "bytes=-3", size: 10, wantStart: 7, wantEnd: 10, wantValid: true},
		{value: "bytes=2-99", size: 10, wantStart: 2, wantEnd: 10, wantValid: true},
		{value: "bytes=10-", size: 10},
	}

	for _, tt := range tests {
		start, end, valid := parseByteRange(tt.value, tt.size)
		if start != tt.wantStart || end != tt.wantEnd || valid != tt.wantValid {
			t.Errorf("parseByteRange(%q, %d) = (%d, %d, %t), want (%d, %d, %t)",
				tt.value, tt.size, start, end, valid, tt.wantStart, tt.wantEnd, tt.wantValid)
		}
	}
}
