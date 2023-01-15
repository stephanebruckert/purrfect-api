package app

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mehanizm/airtable"
	"github.com/pkg/errors"
	"github.com/stephanebruckert/purrfect-api/internal/config"
	"github.com/stephanebruckert/purrfect-api/internal/stats"
	"github.com/stephanebruckert/purrfect-api/internal/webhooks"
	ws "github.com/stephanebruckert/purrfect-api/internal/websocket"
	"log"
)

type App struct {
	Config          *config.Config
	WS              *websocket.Conn
	AirtableRecords []*airtable.Record
	AirtableClient  *airtable.Client
	AirtableBase    *airtable.Base
}

func New() (*App, error) {
	var err error
	app := &App{}

	app.Config, err = config.NewConfig()
	if err != nil {
		return app, errors.Wrap(err, "could not create config")
	}

	client := airtable.NewClient(app.Config.ApiToken)
	app.AirtableClient = client
	bases, err := client.GetBases().WithOffset("").Do()
	if err != nil {
		return app, errors.Wrap(err, "could not get bases")
	}

	for _, base := range bases.Bases {
		if base.Name == app.Config.BaseName {
			app.AirtableBase = base
			break
		}
	}

	return app, err
}

func (app *App) Run() error {
	whooks, err := webhooks.ListWebhooks(app.Config, app.AirtableBase.ID)
	if err != nil {
		return err
	}

	log.Printf("Found %d webhooks\n", len(whooks))

	if len(whooks) > 0 {
		// Delete webhook
		err = webhooks.DeleteWebhook(app.Config, whooks[0].ID, app.AirtableBase.ID)
		if err != nil {
			return err
		}
	}

	// Always create a fresh webhook because they expire after 1 week
	_, err = webhooks.CreateWebhook(app.Config, app.AirtableBase.ID)
	if err != nil {
		return err
	}

	err = app.FetchAirtableData()
	if err != nil {
		return err
	}

	r := app.setupRouter()
	return r.Run("0.0.0.0:3000")
}

func (app *App) FetchAirtableData() error {
	table := app.AirtableClient.GetTable(app.AirtableBase.ID, app.Config.TableName)
	var allRecordsTmp []*airtable.Record

	offset := ""
	for true {
		records, err := table.GetRecords().
			FromView("Grid view").
			WithOffset(offset).
			Do()
		if err != nil {
			return errors.Wrap(err, "could not get records")
		}

		log.Printf("Found %+v at offset %s\n", len(records.Records), records.Offset)
		allRecordsTmp = append(allRecordsTmp, records.Records...)

		if records.Offset == "" {
			break
		}
		offset = records.Offset
	}

	app.AirtableRecords = allRecordsTmp

	log.Printf("Found %d records\n", len(app.AirtableRecords))

	return nil
}

func (app *App) setupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(cors)

	r.GET("/ws", func(c *gin.Context) {
		app.WS = ws.WsEndpoint(c.Writer, c.Request)
	})

	r.POST("/", func(c *gin.Context) {
		err := app.FetchAirtableData()
		if err != nil {
			log.Println(err)
		}
		text := []byte("{}") // Send any valid JSON
		if err := app.WS.WriteMessage(websocket.TextMessage, text); err != nil {
			log.Println(err)
		}
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"health": "OK",
		})
	})

	r.GET("/stats", func(c *gin.Context) {
		stts, err := stats.GetStats(app.AirtableRecords)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"total_orders":      len(app.AirtableRecords),
				"total_cancelled":   stts.TotalCancelled,
				"total_in_progress": stts.TotalInProgress,
				"total_placed":      stts.TotalPlaced,
				"total_shipped":     stts.TotalShipped,
				"total_last_month":  stts.TotalLastMonth,
				"revenue":           stts.Revenue,
				"totals_products":   stts.PerProduct,
			})
		}
	})

	return r
}

func cors(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, "+
		"Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	c.Header("Access-Control-Allow-Origin", "*")
}
