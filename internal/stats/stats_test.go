package stats

import (
	"github.com/mehanizm/airtable"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/maps"
	"testing"
)

func Test_getStats(t *testing.T) {
	t.Run("no rows", func(t *testing.T) {
		var airtableRecords []*airtable.Record
		stats, err := GetStats(airtableRecords)
		assert.Nil(t, err)
		assert.Equal(t, 0.0, stats.Revenue)
		assert.Equal(t, 0, stats.TotalCancelled)
		assert.Equal(t, 0, stats.TotalShipped)
		assert.Equal(t, 0, stats.TotalPlaced)
		assert.Equal(t, 0, stats.TotalInProgress)
		assert.Equal(t, 0, stats.TotalLastMonth)
		assert.Equal(t, 0, len(stats.PerProduct))
	})

	t.Run("rows exist", func(t *testing.T) {
		airtableRecords := []*airtable.Record{
			{
				Fields: map[string]interface{}{
					"product_name": "bow",
					"price":        12412.33,
					"order_status": "in_progress",
					"order_placed": "2006-01-02",
				},
			},
			{
				Fields: map[string]interface{}{
					"product_name": "fish necklace",
					"price":        2.33,
					"order_status": "in_progress",
					"order_placed": "2006-01-02",
				},
			},
			{
				Fields: map[string]interface{}{
					"product_name": "bow",
					"price":        234234234.00,
					"order_status": "cancelled",
					"order_placed": "2006-01-02",
				},
			},
		}
		stats, err := GetStats(airtableRecords)
		assert.Nil(t, err)
		assert.Equal(t, 12414.66, stats.Revenue)
		assert.Equal(t, 1, stats.TotalCancelled)
		assert.Equal(t, 0, stats.TotalShipped)
		assert.Equal(t, 0, stats.TotalPlaced)
		assert.Equal(t, 2, stats.TotalInProgress)
		assert.Equal(t, 0, stats.TotalLastMonth)
		assert.True(t, maps.Equal(map[string]int{
			"bow":           2,
			"fish necklace": 1,
		}, stats.PerProduct))
	})

	t.Run("revenue doesn't include cancelled orders", func(t *testing.T) {
		airtableRecords := []*airtable.Record{
			{
				Fields: map[string]interface{}{
					"product_name": "bow",
					"price":        12412.33,
					"order_status": "cancelled",
					"order_placed": "2006-01-02",
				},
			},
		}
		stats, err := GetStats(airtableRecords)
		assert.Nil(t, err)
		assert.Equal(t, 0.0, stats.Revenue)
	})

	t.Run("date error", func(t *testing.T) {
		airtableRecords := []*airtable.Record{
			{
				Fields: map[string]interface{}{
					"order_placed": "06-01-02",
					"order_status": "shipped",
				},
			},
		}
		_, err := GetStats(airtableRecords)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "incorrect date format: parsing time \"06-01-02\" as \"2006-01-02\": cannot parse \"1-02\" as \"2006\"")
	})

	t.Run("order status error", func(t *testing.T) {
		airtableRecords := []*airtable.Record{
			{
				Fields: map[string]interface{}{
					"order_status": "bad order status",
				},
			},
		}
		_, err := GetStats(airtableRecords)
		assert.NotNil(t, err)
		assert.EqualError(t, err, "unknown order status")
	})
}
