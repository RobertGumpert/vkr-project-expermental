package postService

import (
	"crypto/tls"
	"html/template"
	"testing"
)

func TestTlsConnectionFlow(t *testing.T) {
	_, err := NewTlsAgent(
		"walkmanmail19@gmail.com",
		"QUADRopheniamail12345",
		"",
		465,
		"smtp.gmail.com",
		&tls.Config{
			ServerName: "smtp.gmail.com",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTlsSendLetterTextFlow(t *testing.T) {
	agent, err := NewTlsAgent(
		"walkmanmail19@gmail.com",
		"QUADRopheniamail12345",
		"",
		465,
		"smtp.gmail.com",
		&tls.Config{
			ServerName: "smtp.gmail.com",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	msg := NewMessage(
		agent.ClientBox(),
		"vladislav.kuznetsovRTN1@yandex.ru",
	).Subject(
		"Привет!",
	).Text(
		"Категорически приветсвую!",
	)
	err = agent.SendLetter(msg.GetBytes(), msg.GetReceiver())
	if err != nil {
		t.Fatal(err)
	}
}

func TestTlsSendLetterDynamicHtmlFlow(t *testing.T) {
	type data struct {
		Name  string
		Owner string
		URL   string
	}
	tmpl, err := template.ParseFiles("C:/VKR/vkr-project-expermental/app/data/assets/email-defer-message.html")
	if err != nil {
		t.Fatal(err)
	}
	agent, err := NewTlsAgent(
		"walkmanmail19@gmail.com",
		"QUADRopheniamail12345",
		"",
		465,
		"smtp.gmail.com",
		&tls.Config{
			ServerName: "smtp.gmail.com",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	msg := NewMessage(
		agent.ClientBox(),
		"walkmanmail19@gmail.com",
	).Subject(
		"Суп снял, поставил на балкон. Вот такая хератень будет сваливаться на почту.",
	).DynamicHtml(
		tmpl,
		data{
			Name:  "react",
			Owner: "facebook",
			URL:   "https://metanit.com/go/web/2.1.php",
		},
	)
	err = agent.SendLetter(msg.GetBytes(), msg.GetReceiver())
	if err != nil {
		t.Fatal(err)
	}
}
