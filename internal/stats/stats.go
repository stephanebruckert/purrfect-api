package stats

import (
	"github.com/mehanizm/airtable"
	"github.com/pkg/errors"
	"math"
	"time"
)

type Stats struct {
	TotalCancelled  int
	TotalShipped    int
	TotalPlaced     int
	TotalInProgress int
	TotalLastMonth  int
	Revenue         float64
	PerProduct      map[string]int
}

func GetStats(airtableRecords []*airtable.Record) (Stats, error) {
	stats := Stats{}

	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	stats.PerProduct = map[string]int{}
	for _, record := range airtableRecords {
		// Count by status
		switch record.Fields["order_status"] {
		case "cancelled":
			stats.TotalCancelled++
		case "shipped":
			stats.TotalShipped++
		case "placed":
			stats.TotalPlaced++
		case "in_progress":
			stats.TotalInProgress++
		default:
			return stats, errors.New("unknown order status")
		}

		// Count by date
		input := record.Fields["order_placed"].(string)
		orderTime, err := time.Parse("2006-01-02", input)
		if err != nil {
			return stats, errors.Wrap(err, "incorrect date format")
		}
		if orderTime.After(oneMonthAgo) {
			stats.TotalLastMonth++
		}

		// Sum revenue
		price := record.Fields["price"].(float64)
		if record.Fields["order_status"] != "cancelled" {
			stats.Revenue += price
		}

		// Filter by product
		productName := record.Fields["product_name"].(string)
		_, ok := stats.PerProduct[productName]
		if ok {
			stats.PerProduct[productName]++
		} else {
			stats.PerProduct[productName] = 1
		}
	}

	stats.Revenue = math.Round(stats.Revenue*100) / 100 // Round to closest .00

	return stats, nil
}
