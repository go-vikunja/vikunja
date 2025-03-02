// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
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
func T(lang, key string, params ...string) string {
	translator.mu.RLock()
	defer translator.mu.RUnlock()

	// Try requested language
	if langMap, exists := translator.translations[lang]; exists {
		if translation, found := langMap[key]; found {
			if len(params) > 0 {
				return fmt.Sprintf(translation, stringSliceToInterfaceSlice(params)...)
			}
			return translation
		}
	}

	// Try fallback language if different from requested
	if translator.fallbackLang != lang {
		if langMap, exists := translator.translations[translator.fallbackLang]; exists {
			if translation, found := langMap[key]; found {
				if len(params) > 0 {
					return fmt.Sprintf(translation, stringSliceToInterfaceSlice(params)...)
				}
				return translation
			}
		}
	}

	// Return the key if no translation found
	return key
}

// stringSliceToInterfaceSlice converts a string slice to an interface slice for fmt.Sprintf
func stringSliceToInterfaceSlice(strings []string) []interface{} {
	interfaces := make([]interface{}, len(strings))
	for i, s := range strings {
		interfaces[i] = s
	}
	return interfaces
}

func HasLanguage(lang string) bool {
	_, exists := translator.translations[lang]
	return exists
}
