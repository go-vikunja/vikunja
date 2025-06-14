// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"code.vikunja.io/api/pkg/log"
)

//go:embed lang/*.json
var localeFS embed.FS

// TranslationStore represents a collection of translation entries
type TranslationStore map[string]string

// Translator manages translations for different languages
type Translator struct {
	translations map[string]TranslationStore // language code -> flattened key-value pairs
	fallbackLang string
	mu           sync.RWMutex
}

var translator = &Translator{
	translations: make(map[string]TranslationStore),
	fallbackLang: "en",
}

var availableLanguages = map[string]bool{
	"en":       true,
	"de-DE":    true,
	"de-swiss": true,
	"ru-RU":    true,
	"fr-FR":    true,
	"vi-VN":    true,
	"it-IT":    true,
	"cs-CZ":    true,
	"pl-PL":    true,
	"nl-NL":    true,
	"pt-PT":    true,
	"zh-CN":    true,
	"no-NO":    true,
	"es-ES":    true,
	"da-DK":    true,
	"ja-JP":    true,
	"hu-HU":    true,
	"ar-SA":    true,
	"sl-SI":    true,
	"pt-BR":    true,
	"hr-HR":    true,
	"uk-UA":    true,
	"lt-LT":    true,
	"bg-BG":    true,
	"ko-KR":    true,
	"tr-TR":    true,
	"fi-FI":    true,
	"he-IL":    true,
	// IMPORTANT: Also add new languages to the frontend
}

// Init initializes the global translator with translation files
func Init() {
	dir := "lang"
	entries, err := fs.ReadDir(localeFS, dir)
	if err != nil {
		log.Fatalf("Failed to read embedded translation directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		langCode := strings.TrimSuffix(entry.Name(), ".json")

		if !availableLanguages[langCode] {
			continue
		}

		filePath := filepath.Join(dir, entry.Name())

		err = translator.loadFile(localeFS, langCode, filePath)
		if err != nil {
			log.Fatalf("Failed to load translation file %s: %v", filePath, err)
		}
	}
}

// loadFile loads a translation file for the specified language from the embedded filesystem
func (t *Translator) loadFile(fs embed.FS, langCode, filePath string) error {
	data, err := fs.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var nestedData map[string]interface{}
	if err := json.Unmarshal(data, &nestedData); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	t.mu.Lock()
	// Create or get the flattened map for this language
	if _, exists := t.translations[langCode]; !exists {
		t.translations[langCode] = make(TranslationStore)
	}

	// Flatten the nested structure
	t.flattenTranslations(langCode, nestedData, "")
	t.mu.Unlock()

	return nil
}

// flattenTranslations recursively flattens the nested translation structure
func (t *Translator) flattenTranslations(langCode string, data map[string]interface{}, prefix string) {
	for key, value := range data {
		// Build the full key path
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// If value is a string, add it to the flattened map
		if strValue, ok := value.(string); ok {
			t.translations[langCode][fullKey] = strValue
		} else if mapValue, ok := value.(map[string]interface{}); ok {
			// If value is another map, recurse with the updated prefix
			t.flattenTranslations(langCode, mapValue, fullKey)
		}
	}
}

// GetAvailableLanguages returns a list of available language codes
func GetAvailableLanguages() []string {
	translator.mu.RLock()
	defer translator.mu.RUnlock()

	languages := make([]string, 0, len(translator.translations))
	for lang := range translator.translations {
		languages = append(languages, lang)
	}

	return languages
}

// T returns the translation for the specified key using dot notation in the specified language
func T(lang, key string, params ...any) string {
	translator.mu.RLock()
	defer translator.mu.RUnlock()

	// Try requested language
	if langMap, exists := translator.translations[lang]; exists {
		if translation, found := langMap[key]; found {
			if len(params) > 0 {
				return fmt.Sprintf(translation, params...)
			}
			return translation
		}
	}

	// Try fallback language if different from requested
	if translator.fallbackLang != lang {
		if langMap, exists := translator.translations[translator.fallbackLang]; exists {
			if translation, found := langMap[key]; found {
				if len(params) > 0 {
					return fmt.Sprintf(translation, params...)
				}
				return translation
			}
		}
	}

	// Return the key if no translation found
	return key
}

func HasLanguage(lang string) bool {
	_, exists := translator.translations[lang]
	return exists
}

// TP returns the appropriate pluralized translation string for the specified key, count, and language.
// It expects pluralization rules to be encoded in the translation string using '|' as a separator.
//   - For "singular | plural" (2 forms): uses the first for count 1, second otherwise.
//   - For "zero | one | other" (3 forms): uses the first for count 0, second for count 1, third otherwise.
//
// If the translation string for the key does not contain '|', it's returned as
// If the translation string for the key does not contain '|', it's returned as is.
// If the key is not found, the key itself is returned.
// If the pluralization string is malformed (e.g. contains '|' but not 2 or 3 valid parts), the key is returned and a warning is logged.
// This function does NOT perform any variable interpolation (e.g., replacing "{count}" with the actual number).
func TP(lang, key string, count int64, params ...any) string {
	translator.mu.RLock()
	defer translator.mu.RUnlock()

	var rawTranslation string
	var found bool

	// Try requested language
	if langMap, exists := translator.translations[lang]; exists {
		if translation, keyFound := langMap[key]; keyFound {
			rawTranslation = translation
			found = true
		}
	}

	// Try fallback language if different from requested and not found yet
	if !found && translator.fallbackLang != lang {
		if langMap, exists := translator.translations[translator.fallbackLang]; exists {
			if translation, keyFound := langMap[key]; keyFound {
				rawTranslation = translation
				found = true
			}
		}
	}

	if !found {
		return key // Return the key if no translation found
	}

	// If the string doesn't contain a pipe, it's not a pluralized string according to convention.
	// Return it as is.
	if !strings.Contains(rawTranslation, "|") {
		if len(params) > 0 && strings.Contains(rawTranslation, "%") {
			return fmt.Sprintf(rawTranslation, params...)
		}
		return rawTranslation
	}

	choices := strings.Split(rawTranslation, "|")
	numChoices := len(choices)

	var selectedChoice string

	switch numChoices {
	case 2: // Example: "car | cars" (singular | plural)
		// Handles cases like "1 car" vs "0 cars", "2 cars".
		if count == 1 {
			selectedChoice = choices[0]
		} else {
			selectedChoice = choices[1]
		}
	case 3: // Example: "no apples | one apple | {count} apples" (zero | one | other)
		// Handles cases like "0 apples", "1 apple", "10 apples".
		switch count {
		case 0:
			selectedChoice = choices[0]
		case 1:
			selectedChoice = choices[1]
		default:
			selectedChoice = choices[2]
		}
	default:
		// This case is reached if strings.Contains(rawTranslation, "|") is true,
		// but the number of resulting parts from split is not 2 or 3.
		// This indicates a malformed pluralization string in the translation file.
		log.Errorf("Malformed plural string for key '%s' in lang '%s': %d parts found (expected 2 or 3). Raw string: '%s'", key, lang, numChoices, rawTranslation)
		return key // Return the key to indicate an issue with the translation data.
	}

	selectedChoice = strings.TrimSpace(selectedChoice)

	if len(params) > 0 && strings.Contains(selectedChoice, "%") {
		return fmt.Sprintf(selectedChoice, params...)
	}
	return selectedChoice
}
