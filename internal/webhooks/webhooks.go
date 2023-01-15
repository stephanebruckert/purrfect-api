package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stephanebruckert/purrfect-api/internal/config"
	"io"
	"net/http"
)

type Webhook struct {
	ID string `json:"id"`
}

type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

type CreateWebhookRequest struct {
	NotificationUrl string `json:"notificationUrl"`
	Specification   struct {
		Options struct {
			Filters struct {
				DataTypes []string `json:"dataTypes"`
			} `json:"filters"`
		} `json:"options"`
	} `json:"specification"`
}

func ListWebhooks(cfg *config.Config, baseId string) ([]Webhook, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", baseId), nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making http request")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var webhooks ListWebhooksResponse
	err = json.Unmarshal(body, &webhooks)
	if err != nil {
		return nil, err
	}

	return webhooks.Webhooks, nil
}

func CreateWebhook(cfg *config.Config, baseId string) ([]string, error) {
	cwr := CreateWebhookRequest{}
	cwr.NotificationUrl = cfg.SmeeURL
	cwr.Specification.Options.Filters.DataTypes = []string{"tableData"}

	b, err := json.Marshal(cwr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", baseId),
		bytes.NewBuffer(b))
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
		"Content-Type":  {"application/json"},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error making http request")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var webhooks ListWebhooksResponse
	err = json.Unmarshal(body, &webhooks)
	return nil, err
}

func DeleteWebhook(cfg *config.Config, webhookId string, baseId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks/%s", baseId, webhookId), nil)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "error making http request")
	}

	_, err = io.ReadAll(res.Body)
	return err
}
