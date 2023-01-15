package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stephanebruckert/purrfect-api/internal/config"
	"io"
	"net/http"
	"os"
)

type Webhook struct {
	ID string `json:"id"`
}

type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

func ListWebhooks(cfg *config.Config) ([]Webhook, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", cfg.Base.ID), nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", string(body))

	var webhooks ListWebhooksResponse
	err = json.Unmarshal(body, &webhooks)
	if err != nil {
		return nil, err
	}

	return webhooks.Webhooks, nil
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

func CreateWebhook(cfg *config.Config) ([]string, error) {
	cwr := CreateWebhookRequest{}
	smeeUrl := os.Getenv("SMEE_URL")
	cwr.NotificationUrl = smeeUrl
	cwr.Specification.Options.Filters.DataTypes = []string{"tableData"}
	//cwr.Specification.Options.Filters.RecordChangeScope = "tbltp8DGLhqbUmjK1"

	b, err := json.Marshal(cwr)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", cfg.Base.ID),
		bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
		"Content-Type":  {"application/json"},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", string(body))
	var webhooks ListWebhooksResponse
	err = json.Unmarshal(body, &webhooks)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func DeleteWebhook(webhookId string, cfg *config.Config) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks/%s", cfg.Base.ID, webhookId), nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.ApiToken)},
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("client: error making http request: %s\n", err)
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("%+v\n", string(body))
	return nil
}
