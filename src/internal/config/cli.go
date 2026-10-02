package config

import (
	"flag"
	"fmt"
	"strings"
)

// Usage содержит справку о поддерживаемых параметрах командной строки.
const Usage = `Использование: cmd-emulator [параметры]
  --vfs <путь>      Физический файл VFS (обязателен в CLI или JSON)
  --prompt <текст>  Приглашение REPL (по умолчанию: "$ ")
  --script <путь>   Стартовый скрипт; пустая строка отключает скрипт из JSON
  --config <путь>   Конфигурационный JSON-файл
  --help, -h       Показать справку
Приоритет: CLI → JSON → значения по умолчанию.
Поддерживаются формы --имя значение и --имя=значение.
`

// parseCLI разбирает параметры без позиционных аргументов и повторяющихся флагов.
// Возвращает явно заданные значения либо ошибку; --help и -h возвращают flag.ErrHelp.
func parseCLI(args []string) (overrides, error) {
	const op = "config.parseCLI"
	var cli overrides
	fields := map[string]**string{
		"--vfs": &cli.vfs, "--prompt": &cli.prompt,
		"--script": &cli.script, "--config": &cli.config,
	}
	for i := 0; i < len(args); i++ {
		if args[i] == "--help" || args[i] == "-h" {
			return overrides{}, flag.ErrHelp
		}
		name, value, inline := strings.Cut(args[i], "=")
		field, ok := fields[name]
		if !ok {
			return overrides{}, fmt.Errorf("%s: неизвестный параметр %q", op, args[i])
		}
		if *field != nil {
			return overrides{}, fmt.Errorf("%s: параметр %s указан повторно", op, name)
		}
		if !inline {
			if i+1 == len(args) || strings.HasPrefix(args[i+1], "--") || args[i+1] == "-h" {
				return overrides{}, fmt.Errorf("%s: после %s требуется значение", op, name)
			}
			i++
			value = args[i]
		}
		*field = &value
	}
	return cli, nil
}
