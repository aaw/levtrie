package levtrie

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
)

var data []string
var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
var words []string
var suggestData = []string{
	"acetonylacetone",
	"barbaralalia",
	"calcic",
	"dark",
	"using",
	"volt",
	"wrenchingly",
	"xenos",
	"yore",
	"zymosis",
}

func randString(n int) string {
	runes := make([]rune, n)
	for i := range runes {
		runes[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(runes)
}

func ensureData(n int) {
	if n <= len(data) {
		return
	}
	for i := len(data); i < n; i++ {
		data = append(data, randString(rand.Intn(20)))
	}
}

func ensureWords() {
	if len(words) > 0 {
		return
	}
	filename := "/usr/share/dict/words"
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("%v: %v", filename, err))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		word := strings.ToLower(scanner.Text())
		words = append(words, word)
	}
}

func benchmarkSuggest(d int, b *testing.B) {
	ensureWords()
	r := New()
	for _, word := range words {
		r.Set(word, word)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Suggest(suggestData[i%len(suggestData)], int8(d), 10)
	}
}

func benchmarkSuggestAfterExactPrefix(d int, p int, b *testing.B) {
	ensureWords()
	r := New()
	for _, word := range words {
		r.Set(word, word)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.SuggestAfterExactPrefix(suggestData[i%len(suggestData)], p, int8(d), 10)
	}
}

func BenchmarkSuggestTopTenDistance1(b *testing.B) {
	benchmarkSuggest(1, b)
}

func BenchmarkSuggestTopTenDistance2(b *testing.B) {
	benchmarkSuggest(2, b)
}

func BenchmarkSuggestTopTenDistance3(b *testing.B) {
	benchmarkSuggest(3, b)
}

func BenchmarkSuggestTopTenDistance4(b *testing.B) {
	benchmarkSuggest(4, b)
}

func BenchmarkSuggestTopTenDistance5(b *testing.B) {
	benchmarkSuggest(5, b)
}

func BenchmarkSuggestTopTenDistance6(b *testing.B) {
	benchmarkSuggest(6, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance1(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(1, 1, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance2(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(2, 1, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance3(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(3, 1, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance4(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(4, 1, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance5(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(5, 1, b)
}

func BenchmarkSuggestAfterLength1PrefixTopTenDistance6(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(6, 1, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance1(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(1, 2, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance2(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(2, 2, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance3(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(3, 2, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance4(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(4, 2, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance5(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(5, 2, b)
}

func BenchmarkSuggestAfterLength2PrefixTopTenDistance6(b *testing.B) {
	benchmarkSuggestAfterExactPrefix(6, 2, b)
}

func BenchmarkLevtrieSet(b *testing.B) {
	ensureData(b.N)
	b.ResetTimer()
	r := New()
	for i := 0; i < b.N; i++ {
		r.Set(data[i], data[i])
	}
}

func BenchmarkMapSet(b *testing.B) {
	ensureData(b.N)
	b.ResetTimer()
	m := make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[data[i]] = data[i]
	}
}

func BenchmarkLevtrieGet(b *testing.B) {
	ensureData(b.N)
	r := New()
	for i := 0; i < b.N; i++ {
		r.Set(data[i], data[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Get(data[i])
	}
}

func BenchmarkMapGet(b *testing.B) {
	ensureData(b.N)
	m := make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[data[i]] = data[i]
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m[data[i]]
	}
}

func BenchmarkLevtrieDelete(b *testing.B) {
	ensureData(b.N)
	r := New()
	for i := 0; i < b.N; i++ {
		r.Set(data[i], data[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Delete(data[i])
	}
}

func BenchmarkMapDelete(b *testing.B) {
	ensureData(b.N)
	m := make(map[string]string)
	for i := 0; i < b.N; i++ {
		m[data[i]] = data[i]
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		delete(m, data[i])
	}
}
