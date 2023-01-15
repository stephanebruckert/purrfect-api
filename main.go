package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mehanizm/airtable"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"os"
)

const (
	BaseName  = "Purrfect Creations (Copy)"
	TableName = "Orders"
)

type Config struct {
	airtableClient *airtable.Client
	base           *airtable.Base
	apiToken       string
}

var allRecords []*airtable.Record

func getCancelledRecords() []*airtable.Record {
	var cancelledRecords []*airtable.Record

	for _, record := range allRecords {
		if record.Fields["order_status"] == "cancelled" {
			cancelledRecords = append(cancelledRecords, record)
		}
	}

	return cancelledRecords
}

func cors(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	c.Header("Access-Control-Allow-Origin", "*")
}

func (cfg Config) setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors)

	r.POST("/", func(c *gin.Context) {
		err := cfg.Init()
		if err != nil {
			fmt.Println(err)
		}
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"health": "OK",
		})
	})

	r.GET("/totals", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"total_orders":    len(allRecords),
			"total_cancelled": len(getCancelledRecords()),
		})
	})

	return r
}

func (cfg Config) Init() error {
	table := cfg.airtableClient.GetTable(cfg.base.ID, TableName)

	offset := ""
	for true {
		records, err := table.GetRecords().
			FromView("Grid view").
			WithOffset(offset).
			Do()
		if err != nil {
			return errors.Wrap(err, "Could not get records")
		}

		fmt.Printf("Found %+v at offset %s\n", len(records.Records), records.Offset)
		allRecords = append(allRecords, records.Records...)

		if records.Offset == "" {
			break
		}
		offset = records.Offset
	}

	fmt.Printf("Found %+v\n", len(allRecords))

	return nil
}

func NewConfig() (Config, error) {
	cfg := Config{}
	apiToken := os.Getenv("AIRTABLE_API_TOKEN")
	client := airtable.NewClient(apiToken)
	cfg.airtableClient = client
	bases, err := client.GetBases().WithOffset("").Do()
	if err != nil {
		return cfg, errors.Wrap(err, "Could not get bases")
	}

	for _, base := range bases.Bases {
		if base.Name == BaseName {
			cfg.base = base
			break
		}
	}
	return cfg, nil
}

type Webhook struct {
	ID string `json:"id"`
}

type ListWebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

func (cfg Config) ListWebhooks() ([]Webhook, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", cfg.base.ID), nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.apiToken)},
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

func (cfg Config) CreateWebhook() ([]string, error) {
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
		fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks", cfg.base.ID),
		bytes.NewBuffer(b))
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return nil, err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.apiToken)},
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

func (cfg Config) DeleteWebhook(webhookId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://api.airtable.com/v0/bases/%s/webhooks/%s", cfg.base.ID, webhookId), nil)
	if err != nil {
		fmt.Printf("client: could not create request: %s\n", err)
		return err
	}

	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", cfg.apiToken)},
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

func main() {
	cfg, err := NewConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	apiToken := os.Getenv("AIRTABLE_API_TOKEN")
	cfg.apiToken = apiToken

	webhooks, err := cfg.ListWebhooks()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Found %d webhooks\n", len(webhooks))

	if len(webhooks) > 0 {
		// Delete webhook
		err = cfg.DeleteWebhook(webhooks[0].ID)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Println("Always create a fresh webhook because they expire after 1 week")
	_, err = cfg.CreateWebhook()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = cfg.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	r := cfg.setupRouter()
	err = r.Run("localhost:3000")
	if err != nil {
		fmt.Println(err)
	}
}
