// Package config загружает и проверяет настройки запуска независимо от GUI.
package config

import (
	"fmt"
	"path/filepath"
)

// DefaultPrompt используется, если приглашение не задано в JSON или CLI.
const DefaultPrompt = "$ "

// Settings содержит итоговые настройки после применения приоритетов и проверки файлов.
type Settings struct {
	VFSPath    string
	Prompt     string
	ScriptPath string
	ConfigPath string
}

// overrides хранит отдельно отсутствие параметра и явно переданное пустое значение.
type overrides struct {
	vfs    *string
	prompt *string
	script *string
	config *string
}

// Load читает CLI и JSON, применяет приоритет CLI → JSON → значения по умолчанию.
// Возвращает проверенные настройки либо ошибку; справка определяется через errors.Is(err, flag.ErrHelp).
func Load(args []string) (Settings, error) {
	const op = "config.Load"
	cli, err := parseCLI(args)
	if err != nil {
		return Settings{}, fmt.Errorf("%s: %w", op, err)
	}
	settings := Settings{Prompt: DefaultPrompt}
	if cli.config != nil {
		if *cli.config == "" {
			return Settings{}, fmt.Errorf("%s: --config требует непустой путь", op)
		}
		settings, err = readJSON(*cli.config)
		if err != nil {
			return Settings{}, fmt.Errorf("%s: %w", op, err)
		}
	}
	applyCLI(&settings, cli)
	if err := settings.validate(); err != nil {
		return Settings{}, fmt.Errorf("%s: %w", op, err)
	}
	return settings, nil
}

// applyCLI заменяет только явно переданные поля настроек, включая пустые prompt и script.
// Изменяет settings и ничего не возвращает.
func applyCLI(settings *Settings, cli overrides) {
	if cli.vfs != nil {
		settings.VFSPath = *cli.vfs
	}
	if cli.prompt != nil {
		settings.Prompt = *cli.prompt
	}
	if cli.script != nil {
		settings.ScriptPath = *cli.script
	}
}

// String возвращает отладочное представление всех итоговых настроек с экранированием строк.
func (s Settings) String() string {
	return fmt.Sprintf("config=%q vfs=%q prompt=%q script=%q", s.ConfigPath, s.VFSPath, s.Prompt, s.ScriptPath)
}

// VFSName возвращает имя физического файла VFS для заголовка окна на этапе 2.
// Содержимое XML и собственное имя VFS будут обрабатываться на этапе 3.
func (s Settings) VFSName() string {
	return filepath.Base(s.VFSPath)
}
