package dashboard

import (
	"embed"
	"encoding/json"
	"path"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

type Locale struct {
	Code    string
	strings map[string]string
}

type LanguageOption struct {
	Code  string
	Label string
}

var supportedLocales = map[string]struct{}{
	"en": {},
	"it": {},
}

var (
	localeLoadOnce sync.Once
	localeCache    map[string]map[string]string
)

var languageOptions = []LanguageOption{
	{Code: "en", Label: "language.option.en"},
	{Code: "it", Label: "language.option.it"},
}

func LoadLocale(code string) Locale {
	code = normalizeLocaleCode(code)
	if _, ok := supportedLocales[code]; !ok {
		code = "en"
	}

	return Locale{Code: code, strings: loadLocaleStrings(code)}
}

func (locale Locale) Text(key string) string {
	if locale.strings == nil {
		return key
	}
	if value, ok := locale.strings[key]; ok {
		return value
	}
	return key
}

func (locale Locale) LanguageOptions() []LanguageOption {
	result := make([]LanguageOption, 0, len(languageOptions))
	for _, option := range languageOptions {
		result = append(result, LanguageOption{Code: option.Code, Label: locale.Text(option.Label)})
	}
	return result
}

func localizedPath(p string, locale Locale) string {
	code := normalizeLocaleCode(locale.Code)
	if code == "en" {
		return p
	}
	separator := "?"
	if strings.Contains(p, "?") {
		separator = "&"
	}
	return p + separator + "lang=" + code
}

func normalizeLocaleCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))
	if idx := strings.Index(code, "-"); idx >= 0 {
		code = code[:idx]
	}
	if idx := strings.Index(code, "_"); idx >= 0 {
		code = code[:idx]
	}
	return code
}

func loadLocaleStrings(code string) map[string]string {
	localeLoadOnce.Do(preloadLocaleCache)
	if strings, ok := localeCache[code]; ok {
		return strings
	}
	return map[string]string{}
}

func preloadLocaleCache() {
	cache := make(map[string]map[string]string, len(supportedLocales))
	for code := range supportedLocales {
		cache[code] = readLocaleFile(code)
	}
	localeCache = cache
}

func readLocaleFile(code string) map[string]string {
	fileName := path.Join("locales", code+".json")
	data, err := localeFS.ReadFile(fileName)
	if err != nil {
		return map[string]string{}
	}

	localeStrings := map[string]string{}
	if err := json.Unmarshal(data, &localeStrings); err != nil {
		return map[string]string{}
	}
	return localeStrings
}
