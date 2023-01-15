package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mehanizm/airtable"
	"github.com/pkg/errors"
	"github.com/stephanebruckert/purrfect-api/internal/config"
	"github.com/stephanebruckert/purrfect-api/internal/webhooks"
	"log"
	"net/http"

	"os"
)

const (
	TableName = "Orders"
)

var WS *websocket.Conn
var allRecords []*airtable.Record

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	WS, _ = upgrader.Upgrade(w, r, nil)

	log.Println("Client Connected")
}

type App struct {
	Config *config.Config
}

func New() (*App, error) {
	var err error
	app := &App{}

	app.Config, err = config.NewConfig()
	if err != nil {
		return app, err
	}
	apiToken := os.Getenv("AIRTABLE_API_TOKEN")
	app.Config.ApiToken = apiToken
	return app, nil
}

func (app App) Run() error {
	whooks, err := webhooks.ListWebhooks(app.Config)
	if err != nil {
		return err
	}

	fmt.Printf("Found %d webhooks\n", len(whooks))

	if len(whooks) > 0 {
		// Delete webhook
		err = webhooks.DeleteWebhook(whooks[0].ID, app.Config)
		if err != nil {
			return err
		}
	}

	fmt.Println("Always create a fresh webhook because they expire after 1 week")
	_, err = webhooks.CreateWebhook(app.Config)
	if err != nil {
		return err
	}

	err = app.Init()
	if err != nil {
		return err
	}

	r := app.setupRouter()
	return r.Run("localhost:3000")
}

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
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, "+
		"Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	c.Header("Access-Control-Allow-Origin", "*")
}

func (app App) setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors)

	r.GET("/ws", func(c *gin.Context) {
		WsEndpoint(c.Writer, c.Request)
	})

	r.POST("/", func(c *gin.Context) {
		err := app.Init()
		if err != nil {
			fmt.Println(err)
		}
		text := []byte("{}")
		if err := WS.WriteMessage(websocket.TextMessage, text); err != nil {
			log.Println(err)
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

func (app App) Init() error {
	table := app.Config.AirtableClient.GetTable(app.Config.Base.ID, TableName)
	var allRecordsTmp []*airtable.Record

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
		allRecordsTmp = append(allRecordsTmp, records.Records...)

		if records.Offset == "" {
			break
		}
		offset = records.Offset
	}

	allRecords = allRecordsTmp

	fmt.Printf("Found %+v\n", len(allRecords))

	return nil
}
