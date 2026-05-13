package util_test

import (
	"testing"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/util"
)

func TestCalculateCRC32C(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "hello world",
			input: "hello world",
		},
		{
			name:  "single character",
			input: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.CalculateCRC32C([]byte(tt.input))
			if result == "" {
				t.Errorf("CRC32C(%q) returned empty string", tt.input)
			}
			if len(result) == 0 {
				t.Errorf("CRC32C(%q) returned empty result", tt.input)
			}
		})
	}
}

func TestCalculateMD5(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "hello world",
			input: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.CalculateMD5([]byte(tt.input))
			if result == "" {
				t.Errorf("MD5(%q) returned empty string", tt.input)
			}
		})
	}
}

func TestChecksumsAreDeterministic(t *testing.T) {
	data := []byte("test data for checksum")

	crc1 := util.CalculateCRC32C(data)
	crc2 := util.CalculateCRC32C(data)
	if crc1 != crc2 {
		t.Errorf("CRC32C not deterministic: %q != %q", crc1, crc2)
	}

	md51 := util.CalculateMD5(data)
	md52 := util.CalculateMD5(data)
	if md51 != md52 {
		t.Errorf("MD5 not deterministic: %q != %q", md51, md52)
	}
}

func TestChecksumsDifferForDifferentData(t *testing.T) {
	crc1 := util.CalculateCRC32C([]byte("data1"))
	crc2 := util.CalculateCRC32C([]byte("data2"))
	if crc1 == crc2 {
		t.Errorf("CRC32C should differ for different data")
	}

	md51 := util.CalculateMD5([]byte("data1"))
	md52 := util.CalculateMD5([]byte("data2"))
	if md51 == md52 {
		t.Errorf("MD5 should differ for different data")
	}
}

func TestCRC32CKnownValue(t *testing.T) {
	data := []byte("123456789")
	result := util.CalculateCRC32C(data)
	expected := "4waSgw=="
	if result != expected {
		t.Errorf("CRC32C('123456789') = %q, want %q", result, expected)
	}
}

func TestMD5KnownValue(t *testing.T) {
	data := []byte("hello world")
	result := util.CalculateMD5(data)
	expected := "XrY7u+Ae7tCTyyK7j1rNww=="
	if result != expected {
		t.Errorf("MD5('hello world') = %q, want %q", result, expected)
	}
}
