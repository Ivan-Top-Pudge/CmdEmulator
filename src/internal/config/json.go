package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type optionalString struct {
	value string
	set   bool
}

type jsonSettings struct {
	VFS    optionalString `json:"vfs"`
	Prompt optionalString `json:"prompt"`
	Script optionalString `json:"script"`
}

// UnmarshalJSON читает необязательное строковое поле и запоминает его присутствие.
// Возвращает ошибку для null или другого типа; успешный разбор возвращает nil.
func (s *optionalString) UnmarshalJSON(data []byte) error {
	const op = "config.optionalString.UnmarshalJSON"
	if strings.TrimSpace(string(data)) == "null" {
		return fmt.Errorf("%s: значение должно быть строкой, а не null", op)
	}
	if err := json.Unmarshal(data, &s.value); err != nil {
		return fmt.Errorf("%s: ожидается строка: %w", op, err)
	}
	s.set = true
	return nil
}

// readJSON читает один JSON-объект с полями vfs, prompt и script без неизвестных полей.
// Возвращает настройки с путями относительно каталога JSON либо ошибку чтения или формата.
func readJSON(path string) (Settings, error) {
	const op = "config.readJSON"
	file, err := os.Open(path)
	if err != nil {
		return Settings{}, fmt.Errorf("%s: открыть конфигурацию %q: %w", op, path, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var fields *jsonSettings
	if err := decoder.Decode(&fields); err != nil {
		return Settings{}, fmt.Errorf("%s: неверный JSON в %q: %w", op, path, err)
	}
	if fields == nil {
		return Settings{}, fmt.Errorf("%s: конфигурация должна быть JSON-объектом", op)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return Settings{}, fmt.Errorf("%s: после JSON-объекта в %q обнаружены лишние данные", op, path)
	} else if err != io.EOF {
		return Settings{}, fmt.Errorf("%s: лишние данные в %q: %w", op, path, err)
	}
	settings := Settings{Prompt: DefaultPrompt, ConfigPath: path}
	settings.VFSPath = resolvePath(path, fields.VFS.value)
	settings.ScriptPath = resolvePath(path, fields.Script.value)
	if fields.Prompt.set {
		settings.Prompt = fields.Prompt.value
	}
	return settings, nil
}

// resolvePath возвращает непустой относительный путь относительно каталога конфигурации.
// Абсолютный путь и пустое значение возвращает без изменений.
func resolvePath(configPath, value string) string {
	if value == "" || filepath.IsAbs(value) {
		return value
	}
	return filepath.Join(filepath.Dir(configPath), value)
}
