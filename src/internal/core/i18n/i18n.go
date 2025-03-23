package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultLang = "ru"
	instance    *I18n
	once        sync.Once
)

type I18n struct {
	currentLang  string
	translations map[string]map[string]string // lang -> key -> text
	mu           sync.RWMutex
}

func GetInstance() *I18n {
	once.Do(func() {
		instance = &I18n{
			currentLang:  defaultLang,
			translations: make(map[string]map[string]string),
		}
	})
	return instance
}

func (i *I18n) LoadTranslations(dir string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read translations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		lang := file.Name()[:len(file.Name())-5]
		path := filepath.Join(dir, file.Name())

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read translation file %s: %w", file.Name(), err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return fmt.Errorf("failed to parse translation file %s: %w", file.Name(), err)
		}

		i.translations[lang] = translations
	}

	return nil
}

func (i *I18n) SetLanguage(lang string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.translations[lang]; ok {
		i.currentLang = lang
	}
}

func (i *I18n) T(key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if trans, ok := i.translations[i.currentLang][key]; ok {
		if len(args) > 0 {
			return fmt.Sprintf(trans, args...)
		}
		return trans
	}

	if i.currentLang != defaultLang {
		if trans, ok := i.translations[defaultLang][key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(trans, args...)
			}
			return trans
		}
	}

	return key
}
