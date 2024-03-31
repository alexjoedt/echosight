package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
)

var (
	ErrNoTelegramBotToken error = errors.New("no telegram bot token")
	ErrNoTelegramChatID   error = errors.New("no telegram chat ID")
)

var _ Sender = (*telegram)(nil)

type telegram struct {
	service  echosight.PreferenceService
	botToken string
	chatIDs  []string
	enabled  bool
}

func NewTelegramBot(prefService echosight.PreferenceService) (*telegram, error) {
	if prefService == nil {
		return nil, fmt.Errorf("preference service is nil")
	}

	t := &telegram{
		service: prefService,
	}

	err := t.loadConfig()
	if err != nil {
		if !errors.Is(err, ErrNoTelegramBotToken) &&
			!errors.Is(err, ErrNoTelegramChatID) {
			return nil, err
		}
	}

	return t, nil
}

func (t *telegram) send(chatID string, text string) error {

	var payload struct {
		Text   string `json:"text"`
		ChatID string `json:"chat_id"`
	}

	payload.ChatID = chatID
	payload.Text = text

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	reader := strings.NewReader(string(data))
	client := &http.Client{} // TODO: timeout
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return nil
}

func (t *telegram) Send(ctx context.Context, result *echosight.Result) error {
	err := t.loadConfig()
	if err != nil {
		return err
	}

	var sendErr []error
	for _, id := range t.chatIDs {
		if err := t.send(id, "TODO: result"); err != nil {
			sendErr = append(sendErr, err)
		}
	}

	return errors.Join(sendErr...)
}

func (t *telegram) Enabled() bool {
	t.loadConfig()
	return t.enabled
}

func (t *telegram) loadConfig() error {
	f := filter.NewDefaultPreferenceFilter()
	f.Name = "telegram"
	rawConfig, err := t.service.List(context.Background(), f)
	if err != nil {
		return err
	}

	if !rawConfig.Has("telegram_bot_token") {
		return ErrNoTelegramBotToken
	}

	if !rawConfig.Has("telegram_chat_ids") {
		return ErrNoTelegramChatID
	}

	t.botToken = rawConfig.Get("telegram_bot_token")
	t.chatIDs = strings.Split(rawConfig.Get("telegram_chat_ids"), ",")
	t.enabled = rawConfig.Get("telegram_enabled") == "true"

	return nil
}
