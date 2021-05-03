package postService

import (
	"bytes"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"html/template"
	"strings"
)

type message struct {
	str              string
	receiver, author string
}

func NewMessage(author, receiver string) *message {
	m := new(message)
	m.receiver = receiver
	m.author = author
	m.str = strings.Join([]string{
		strings.Join([]string{
			"FROM: ", author,
		}, ""),
		strings.Join([]string{
			"TO: ", receiver,
		}, ""),
	}, "\r\n")
	return m
}

func (m *message) Subject(subject string) *message {
	m.str = strings.Join([]string{
		m.str,
		strings.Join([]string{
			"Subject: ", subject,
		}, ""),
	}, "\r\n")
	return m
}

func (m *message) Text(text string) *message {
	m.str = strings.Join([]string{
		m.str,
		strings.Join([]string{
			"Content-type: text/html;charset=utf-8\r\nMIME-Version: 1.0",
			text,
		}, "\r\n"),
	}, "\r\n")
	return m
}

func (m *message) DynamicHtml(template *template.Template, data interface{}) *message {
	buffer := new(bytes.Buffer)
	err := template.Execute(buffer, data)
	if err != nil {
		runtimeinfo.LogError(err)
		return m
	}
	m.str = strings.Join([]string{
		m.str,
		strings.Join([]string{
			"Content-type: text/html;charset=utf-8\r\nMIME-Version: 1.0",
			string(buffer.Bytes()),
		}, "\r\n"),
	}, "\r\n")
	return m
}

func (m *message) GetString() string {
	return m.str
}

func (m *message) GetBytes() []byte {
	return []byte(m.str)
}

func (m *message) GetReceiver() string {
	return m.receiver
}

func (m *message) GetAuthor() string {
	return m.author
}
