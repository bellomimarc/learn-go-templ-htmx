package dashboard

import (
	"sync"
	"testing"
)

var (
	benchLocale Locale
	benchMap    map[string]string
	benchText   string
)

func resetLocaleCacheForBenchmark() {
	localeLoadOnce = sync.Once{}
	localeCache = nil
}

func BenchmarkLoadLocaleCached(b *testing.B) {
	resetLocaleCacheForBenchmark()
	benchLocale = LoadLocale("it")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchLocale = LoadLocale("it")
	}
}

func BenchmarkLoadLocaleFirstCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resetLocaleCacheForBenchmark()
		benchLocale = LoadLocale("it")
	}
}

func BenchmarkLoadLocaleStringsCached(b *testing.B) {
	resetLocaleCacheForBenchmark()
	_ = LoadLocale("it")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchMap = loadLocaleStrings("it")
	}
}

func BenchmarkReadLocaleFileNoCache(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchMap = readLocaleFile("it")
	}
}

func BenchmarkLocaleTextLookup(b *testing.B) {
	benchLocale = LoadLocale("it")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		benchText = benchLocale.Text("loan.submit")
	}
}
