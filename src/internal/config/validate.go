package config

import (
	"fmt"
	"os"
	"strings"
)

// validate проверяет обязательный путь VFS и доступность итоговых файлов после объединения настроек.
// Возвращает nil при успехе либо ошибку; содержимое XML на этапе 2 не разбирает.
func (s Settings) validate() error {
	const op = "config.Settings.validate"
	if strings.TrimSpace(s.VFSPath) == "" {
		return fmt.Errorf("%s: укажите путь к VFS через --vfs или поле vfs в JSON", op)
	}
	if err := validateFile("VFS", s.VFSPath); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if s.ScriptPath != "" {
		if err := validateFile("стартовый скрипт", s.ScriptPath); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

// validateFile проверяет, что путь открывается для чтения и указывает на обычный файл.
// Возвращает nil при успехе либо ошибку с назначением файла и именем операции.
func validateFile(kind, path string) error {
	const op = "config.validateFile"
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%s: открыть %s %q: %w", op, kind, path, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("%s: проверить %s %q: %w", op, kind, path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: %s %q должен быть обычным файлом", op, kind, path)
	}
	return nil
}
