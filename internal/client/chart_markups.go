package client

import (
	"context"

	"github.com/major/marketsurge-agent/internal/models"
	"github.com/major/marketsurge-agent/queries"
)

// GetChartMarkups returns saved chart markups for a symbol.
func (c *Client) GetChartMarkups(ctx context.Context, symbol, frequency, sortDir string) (*models.ChartMarkupList, error) {
	query, err := queries.Load("chart_markups.graphql")
	if err != nil {
		return nil, err
	}

	raw, err := c.Execute(ctx, Request{
		OperationName: "FetchChartMarkups",
		Variables: map[string]any{
			"site":        "marketsurge",
			"dowJonesKey": symbol,
			"frequency":   frequency,
			"sortDir":     sortDir,
		},
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	container := getNestedMap(raw, "data", "user", "chartMarkups")
	result := &models.ChartMarkupList{CursorID: stringify(container["cursorId"]), Markups: []models.ChartMarkup{}}
	for _, entry := range getNestedSlice(container, "chartMarkups") {
		item, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		result.Markups = append(result.Markups, models.ChartMarkup{
			ID:        stringify(item["id"]),
			Name:      stringPtr(item["name"]),
			Data:      stringify(item["data"]),
			Frequency: stringPtr(item["frequency"]),
			Site:      stringPtr(item["site"]),
			CreatedAt: stringPtr(item["createdAt"]),
			UpdatedAt: stringPtr(item["updatedAt"]),
		})
	}
	return result, nil
}
