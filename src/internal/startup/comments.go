package startup

import "strings"

// commentFilter хранит начало открытого блока и состояние текущей строки стартового скрипта.
type commentFilter struct {
	blockLine int
	quote     byte
	escaped   bool
	text      strings.Builder
}

// strip удаляет комментарии из строки, сохраняя кавычки и экранирование для парсера команд.
// Состояние блочного комментария переносится на следующую строку; блок заменяется пробелом.
func (f *commentFilter) strip(line string, lineNumber int) string {
	f.quote, f.escaped = 0, false
	f.text.Reset()
	for i := 0; i < len(line); i++ {
		i = f.consume(line, i, lineNumber)
	}
	return f.text.String()
}

// consume обрабатывает символ с учётом блока, кавычек и экранирования; возвращает последний прочитанный индекс.
func (f *commentFilter) consume(line string, i, lineNumber int) int {
	if f.blockLine != 0 {
		return f.consumeBlock(line, i)
	}
	ch := line[i]
	if f.escaped {
		f.text.WriteByte(ch)
		f.escaped = false
		return i
	}
	if f.quote == '\'' {
		f.writeQuoted(ch)
		return i
	}
	switch ch {
	case '\\':
		f.text.WriteByte(ch)
		f.escaped = true
	case '\'', '"':
		f.writeQuoted(ch)
	default:
		return f.consumeText(line, i, lineNumber)
	}
	return i
}

// writeQuoted добавляет символ и открывает либо закрывает совпадающую кавычку; ничего не возвращает.
func (f *commentFilter) writeQuoted(ch byte) {
	f.text.WriteByte(ch)
	if f.quote == 0 {
		f.quote = ch
	} else if f.quote == ch {
		f.quote = 0
	}
}

// consumeText распознаёт // и /* вне кавычек либо копирует символ; возвращает последний прочитанный индекс.
func (f *commentFilter) consumeText(line string, i, lineNumber int) int {
	if f.quote == 0 && strings.HasPrefix(line[i:], "//") {
		return len(line)
	}
	if f.quote == 0 && strings.HasPrefix(line[i:], "/*") {
		f.blockLine = lineNumber
		f.text.WriteByte(' ')
		return i + 1
	}
	f.text.WriteByte(line[i])
	return i
}

// consumeBlock пропускает текст до первого */ без вложенности и возвращает последний прочитанный индекс.
func (f *commentFilter) consumeBlock(line string, i int) int {
	end := strings.Index(line[i:], "*/")
	if end < 0 {
		return len(line)
	}
	f.blockLine = 0
	return i + end + 1
}
