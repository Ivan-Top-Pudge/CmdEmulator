package shell

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

type parser struct {
	args    []string
	word    strings.Builder
	quote   rune
	escaped bool
	started bool
}

// parse разбирает ввод с учётом кавычек, экранирования и переменных окружения.
// Возвращает команду с аргументами либо nil и ошибку при неверном синтаксисе.
func parse(input string) ([]string, error) {
	const op = "shell.parse"

	p := &parser{}
	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		var err error
		i, err = p.consume(runes, i)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	if p.escaped {
		return nil, fmt.Errorf("%s: после обратного слеша ожидается символ", op)
	}
	if p.quote != 0 {
		return nil, fmt.Errorf("%s: незакрытая кавычка %c", op, p.quote)
	}
	p.finishWord()
	return p.args, nil
}

// consume обрабатывает символ в позиции i и обновляет состояние парсера.
// Возвращает индекс последнего обработанного символа и ошибку раскрытия переменной, если она возникла.
func (p *parser) consume(runes []rune, i int) (int, error) {
	ch := runes[i]
	if p.escaped {
		p.append(ch)
		p.escaped = false
		return i, nil
	}
	if p.quote == '\'' {
		p.consumeSingleQuote(ch)
		return i, nil
	}
	switch ch {
	case '\\':
		p.escaped, p.started = true, true
	case '\'', '"':
		p.consumeQuote(ch)
	case '$':
		value, end, err := expandVariable(runes, i)
		p.word.WriteString(value)
		p.started = true
		return end, err
	default:
		p.consumeText(ch)
	}
	return i, nil
}

// consumeSingleQuote добавляет буквальный символ или закрывает одинарную кавычку.
// Изменяет состояние парсера и ничего не возвращает.
func (p *parser) consumeSingleQuote(ch rune) {
	if ch == '\'' {
		p.quote = 0
		return
	}
	p.append(ch)
}

// consumeQuote открывает или закрывает кавычку, а внутри других кавычек добавляет её как текст.
// Изменяет состояние парсера и ничего не возвращает.
func (p *parser) consumeQuote(ch rune) {
	if p.quote == 0 {
		p.quote, p.started = ch, true
	} else if p.quote == ch {
		p.quote = 0
	} else {
		p.append(ch)
	}
}

// consumeText добавляет символ в аргумент или завершает аргумент на пробеле вне кавычек.
// Изменяет состояние парсера и ничего не возвращает.
func (p *parser) consumeText(ch rune) {
	if p.quote == 0 && unicode.IsSpace(ch) {
		p.finishWord()
		return
	}
	p.append(ch)
}

// append добавляет символ в текущий аргумент и отмечает его начало; ничего не возвращает.
func (p *parser) append(ch rune) {
	p.word.WriteRune(ch)
	p.started = true
}

// finishWord сохраняет начатый аргумент в списке и очищает буфер; ничего не возвращает.
// Если аргумент не начат, оставляет состояние без изменений.
func (p *parser) finishWord() {
	if !p.started {
		return
	}
	p.args = append(p.args, p.word.String())
	p.word.Reset()
	p.started = false
}

// expandVariable раскрывает $NAME или ${NAME}, начиная с символа $ в позиции start.
// Возвращает значение, индекс последнего обработанного символа и ошибку; буквальный $ остаётся текстом.
func expandVariable(runes []rune, start int) (string, int, error) {
	next := start + 1
	if next == len(runes) {
		return "$", start, nil
	}
	if runes[next] == '{' {
		return expandBracedVariable(runes, next)
	}
	if !isVariableStart(runes[next]) {
		return "$", start, nil
	}
	end := next + 1
	for end < len(runes) && isVariablePart(runes[end]) {
		end++
	}
	return os.Getenv(string(runes[next:end])), end - 1, nil
}

// expandBracedVariable проверяет и раскрывает ${NAME}, начиная с открывающей скобки в позиции brace.
// Возвращает значение, индекс закрывающей скобки либо конца ввода и ошибку при неверном синтаксисе.
func expandBracedVariable(runes []rune, brace int) (string, int, error) {
	const op = "shell.expandBracedVariable"

	end := brace + 1
	for end < len(runes) && runes[end] != '}' {
		end++
	}
	if end == len(runes) {
		return "", end, fmt.Errorf("%s: незакрытая переменная окружения", op)
	}
	name := runes[brace+1 : end]
	if len(name) == 0 || !isVariableStart(name[0]) {
		return "", end, fmt.Errorf("%s: неверное имя переменной окружения", op)
	}
	for _, ch := range name {
		if !isVariablePart(ch) {
			return "", end, fmt.Errorf("%s: неверное имя переменной окружения %q", op, string(name))
		}
	}
	return os.Getenv(string(name)), end, nil
}

// isVariableStart возвращает true, если символ может начинать имя переменной: буква ASCII или _.
func isVariableStart(ch rune) bool {
	return ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

// isVariablePart возвращает true, если символ допустим в имени переменной: буква ASCII, цифра или _.
func isVariablePart(ch rune) bool {
	return isVariableStart(ch) || ch >= '0' && ch <= '9'
}
