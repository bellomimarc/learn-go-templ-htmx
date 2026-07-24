package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"
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
	options := []string{"en", "it"}
	sort.Strings(options)

	result := make([]LanguageOption, 0, len(options))
	for _, code := range options {
		result = append(result, LanguageOption{Code: code, Label: locale.Text("language.option." + code)})
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
	return fmt.Sprintf("%s%slang=%s", p, separator, code)
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
	fileName := path.Join("locales", code+".json")
	data, err := localeFS.ReadFile(fileName)
	if err != nil {
		return map[string]string{}
	}

	strings := map[string]string{}
	if err := json.Unmarshal(data, &strings); err != nil {
		return map[string]string{}
	}
	return strings
}
