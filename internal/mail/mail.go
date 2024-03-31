package mail

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"text/template"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/internal/notify"
	"github.com/go-mail/mail"
)

var _ notify.Sender = (*Mailer)(nil)

//go:embed "templates"
var templateFS embed.FS

var (
	ErrWorkerAlreadyRunning error = errors.New("worker is already running")
	ErrMailerNotInitialized error = errors.New("mailer not initialized")
	ErrNoPreferenceService  error = errors.New("no preference service assigned <nil>")
)

type Mailer struct {
	dialer *mail.Dialer
	sender string
	Queue  chan *echosight.Result
	done   chan struct{}
	wg     sync.WaitGroup
	log    *logger.Logger
	Opts
}

type Opts struct {
	AppPreferences echosight.PreferenceService
	Recipients     echosight.RecipientService
	Crypter        echosight.Crypter
}

func New(opts Opts) (*Mailer, error) {

	if opts.AppPreferences == nil {
		return nil, fmt.Errorf("no app preferences in opts")
	}

	if opts.Recipients == nil {
		return nil, fmt.Errorf("no recipient service in opts")
	}

	if opts.Crypter == nil {
		return nil, fmt.Errorf("no crypter in opts")
	}

	workerCount := 3
	m := &Mailer{
		Queue: make(chan *echosight.Result, workerCount),
		done:  make(chan struct{}, 1),
		log:   logger.New("Mailer"),
		wg:    sync.WaitGroup{},
		Opts:  opts,
	}

	// TODO: is this necessary?
	for i := 0; i < workerCount; i++ {
		go m.worker()
	}

	return m, nil
}

func (m *Mailer) Send(ctx context.Context, result *echosight.Result) error {
	m.Queue <- result
	return nil
}

func (m *Mailer) Enabled() bool {
	// TODO: redundant db access...
	config, err := m.getSMTPConfig()
	if err != nil {
		return false
	}

	return config.Enabled
}

func (m *Mailer) getSMTPConfig() (*SMTPConfig, error) {
	if m.AppPreferences == nil {
		return nil, ErrNoPreferenceService
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	f := filter.NewDefaultPreferenceFilter()
	f.Name = "smtp"
	prefs, err := m.AppPreferences.List(ctx, f)
	if err != nil {
		return nil, err
	}

	if len(prefs.Map()) == 0 {
		return nil, fmt.Errorf("no smtp config present")
	}

	var config SMTPConfig
	err = prefs.Decode(&config)
	if err != nil {
		return nil, err
	}

	pass, err := m.Crypter.Decrypt(config.PasswordCrypt)
	if err != nil {
		return nil, err
	}
	config.password = pass
	m.sender = config.Sender

	return &config, nil
}

func (m *Mailer) connect() error {
	config, err := m.getSMTPConfig()
	if err != nil {
		return err
	}

	if !config.Enabled {
		return fmt.Errorf("mail is not enabled")
	}

	portInt, err := strconv.Atoi(config.Port)
	if err != nil {
		return err
	}

	dialer := mail.NewDialer(config.Host, portInt, config.User, config.password)
	dialer.Timeout = time.Second * 1

	cancel, err := dialer.Dial()
	if err != nil {
		return err
	}
	defer cancel.Close()

	dialer.Timeout = time.Second * 5

	m.dialer = dialer
	m.sender = config.Sender

	return nil
}

func (m *Mailer) getRecipients() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	f := filter.NewDefaultRecipientFilter()
	active := true
	f.Active = &active
	result, err := m.Recipients.List(ctx, f)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no recipients")
	}

	recipients := make([]string, len(result))
	for i := range result {
		recipients[i] = result[i].Name
	}

	return recipients, nil
}

func (m *Mailer) SendTemplate(templateFile string, data any) error {
	if err := m.connect(); err != nil {
		return err
	}

	recipients, err := m.getRecipients()
	if err != nil {
		return err
	}

	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipients...)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// three attempts
	for i := 0; i < 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}

		time.Sleep(time.Millisecond * 500)
	}

	return err
}

func (m *Mailer) worker() {
	m.log.Debugf("start worker")
	m.wg.Add(1)
	defer m.wg.Done()

	for {
		select {
		case payload, ok := <-m.Queue:
			if !ok {
				m.log.Debugf("mail channel closed")
				return
			}
			err := m.SendTemplate("state_changed.tmpl", payload)
			if err != nil {
				m.log.Errorf("send mail failed: %v", err)
			} else {
				m.log.Debugf("mail sent")
			}

		case <-m.done:
			m.log.Debugf("mail worker stopped")
			return
		}
	}

}

func (m *Mailer) Stop() error {
	close(m.done)
	m.wg.Wait()
	return nil
}
