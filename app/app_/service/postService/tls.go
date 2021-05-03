package postService

import (
	"crypto/tls"
	"github.com/RobertGumpert/vkr-pckg/runtimeinfo"
	"net/smtp"
	"strconv"
	"strings"
)

type Agent struct {
	clientBox      string
	clientPassword string
	clientIdentity string
	//
	smtpServerDomain     string
	smtpServerTlsPort    int
	smtpServerTlsAddress string
	//
	tlsConfig      *tls.Config
	smtpConnection *tls.Conn
	smtpClient     *smtp.Client
	smtpAuth       smtp.Auth
}

func NewTlsAgent(clientBox, clientPassword, clientIdentity string, smtpServerTlsPort int, smtpServerDomain string, tlsConfig *tls.Config) (*Agent, error) {
	smtpServerTlsAddress := strings.Join(
		[]string{
			smtpServerDomain,
			strconv.Itoa(smtpServerTlsPort),
		},
		":",
	)
	agent := &Agent{
		smtpServerDomain:     smtpServerDomain,
		clientBox:            clientBox,
		smtpServerTlsPort:    smtpServerTlsPort,
		tlsConfig:            tlsConfig,
		smtpServerTlsAddress: smtpServerTlsAddress,
		clientPassword:       clientPassword,
		clientIdentity:       clientIdentity,
	}
	connection, client, err := agent.connect()
	if err != nil {
		return nil, err
	}
	agent.smtpConnection = connection
	agent.smtpClient = client
	return agent, nil
}

func (agent *Agent) connect() (*tls.Conn, *smtp.Client, error) {
	connection, err := tls.Dial(
		"tcp",
		agent.smtpServerTlsAddress,
		agent.tlsConfig,
	)
	if err != nil {
		return nil, nil, err
	}
	client, err := smtp.NewClient(
		connection,
		agent.smtpServerDomain,
	)
	if err != nil {
		return nil, nil, err
	}
	agent.smtpAuth = smtp.PlainAuth(
		agent.clientIdentity,
		agent.clientBox,
		agent.clientPassword,
		agent.smtpServerDomain,
	)
	if err := client.Auth(agent.smtpAuth); err != nil {
		return nil, nil, err
	}
	return connection, client, nil
}

func (agent *Agent) checkRCPTReceivers(receivers ...string) error {
	if err := agent.smtpClient.Mail(agent.clientBox); err != nil {
		return err
	}
	for _, receiver := range receivers {
		if err := agent.smtpClient.Rcpt(receiver); err != nil {
			runtimeinfo.LogError("RCPT err: ", err, " for receiver ", receiver)
		}
	}
	return nil
}

func (agent *Agent) ClientBox() string {
	return agent.clientBox
}

func (agent *Agent) SendLetter(msg []byte, receivers ...string) error {
	if err := agent.checkRCPTReceivers(receivers...); err != nil {
		return err
	}
	writeCloser, err := agent.smtpClient.Data()
	if err != nil {
		return err
	}
	_, err = writeCloser.Write(msg)
	if err != nil {
		return err
	}
	err = writeCloser.Close()
	if err != nil {
		return err
	}
	return nil
}
