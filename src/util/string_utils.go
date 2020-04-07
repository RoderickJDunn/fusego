package util

import "strings"

import "fmt"

// NOTE: TODO: untested
func IndexFrom(s, substr string, offset int) int {
	s_len := len(s)
	if s_len < offset {
		return -1
	}
	if idx := strings.Index(s[offset:], substr); idx >= 0 {
		return int(offset) + idx
	}
	return -1
}

// NOTE: TODO: untested
func LastIndexFrom(s, substr string, offset int) int {
	s_len := len(s)

	if s_len < offset {
		return -1
	}

	patternLen := len(substr)
	// OPTIMIZATION: could be that this cast/cast-back will slow this down...
	runes := []rune(s)

	// Slice string at offset, adding 1 to make substring inclusive (like JS), and add patternLen to allow successful finds when the offset
	//	is in the middle of a substring match. This matches JS V8 behavior
	if s_len < offset+1+patternLen {
		fmt.Println("ERROR: Offset+1+ptn_len is longer than slice bounds")
		fmt.Println("offset", offset)
		fmt.Println("patternLen", patternLen)
		fmt.Println("s", s)
		fmt.Println("substr", substr)
		return -1
	}

	s = string(runes[:offset+1+patternLen])

	// println(s)
	//
	if idx := strings.LastIndex(s, substr); idx >= 0 {
		return idx
	}
	return -1
}
