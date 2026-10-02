package vfs

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type elementXML struct {
	XMLName xml.Name
}

type filesystemXML struct {
	XMLName xml.Name       `xml:"vfs"`
	Name    string         `xml:"name,attr"`
	Roots   []directoryXML `xml:"directory"`
	Extra   []elementXML   `xml:",any"`
	Text    string         `xml:",chardata"`
}

type directoryXML struct {
	Name        string         `xml:"name,attr"`
	Mode        string         `xml:"mode,attr"`
	Directories []directoryXML `xml:"directory"`
	Files       []fileXML      `xml:"file"`
	Extra       []elementXML   `xml:",any"`
	Text        string         `xml:",chardata"`
}

type fileXML struct {
	Name  string       `xml:"name,attr"`
	Mode  string       `xml:"mode,attr"`
	Data  string       `xml:",chardata"`
	Extra []elementXML `xml:",any"`
}

// decode читает ровно один XML-элемент vfs, допуская пробелы, комментарии и XML-декларацию.
// Возвращает описание VFS либо ошибку синтаксиса, корневого элемента или лишних данных.
func decode(data []byte) (filesystemXML, error) {
	const op = "vfs.decode"
	decoder := xml.NewDecoder(bytes.NewReader(data))
	token, err := nextContent(decoder)
	if err != nil {
		return filesystemXML{}, fmt.Errorf("%s: неверный XML: %w", op, err)
	}
	start, ok := token.(xml.StartElement)
	if !ok {
		return filesystemXML{}, fmt.Errorf("%s: ожидается корневой элемент vfs", op)
	}
	var document filesystemXML
	if err := decoder.DecodeElement(&document, &start); err != nil {
		return filesystemXML{}, fmt.Errorf("%s: неверный XML: %w", op, err)
	}
	if _, err := nextContent(decoder); err == nil {
		return filesystemXML{}, fmt.Errorf("%s: после vfs обнаружены лишние данные", op)
	} else if err != io.EOF {
		return filesystemXML{}, fmt.Errorf("%s: неверный XML после vfs: %w", op, err)
	}
	return document, nil
}

// nextContent пропускает XML-комментарии, инструкции и пробельный текст.
// Возвращает следующий значимый токен либо ошибку чтения, включая io.EOF в конце документа.
func nextContent(decoder *xml.Decoder) (xml.Token, error) {
	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch value := token.(type) {
		case xml.Comment, xml.ProcInst:
			continue
		case xml.CharData:
			if strings.TrimSpace(string(value)) == "" {
				continue
			}
		}
		return token, nil
	}
}

// validateContent проверяет отсутствие неизвестных элементов и текста вне содержимого файлов.
// Возвращает nil при успехе либо ошибку формата с именем операции.
func validateContent(extra []elementXML, text string) error {
	const op = "vfs.validateContent"
	if len(extra) != 0 {
		return fmt.Errorf("%s: неизвестный элемент %q", op, extra[0].XMLName.Local)
	}
	if strings.TrimSpace(text) != "" {
		return fmt.Errorf("%s: текст допустим только внутри file", op)
	}
	return nil
}
