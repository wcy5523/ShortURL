package util

import (
	"math"
	"strings"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var base62Map = make(map[rune]int)

func init() {
	for i, c := range base62Chars {
		base62Map[c] = i
	}
}

func ToBase62(num uint64) string {
	if num == 0 {
		return string(base62Chars[0])
	}
	var builder strings.Builder
	for num > 0 {
		remainder := num % 62
		builder.WriteByte(base62Chars[remainder])
		num = num / 62
	}
	return reverse(builder.String())
}

func FromBase62(s string) uint64 {
	var num uint64
	length := len(s)
	for i, c := range s {
		power := length - i - 1
		num += uint64(base62Map[c]) * uint64(math.Pow(62, float64(power)))
	}
	return num
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
