package client

import (
	"context"

	"github.com/tidwall/gjson"

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

	container := gjson.GetBytes(raw, "data.user.chartMarkups")
	return &models.ChartMarkupList{
		CursorID: stringify(container.Get("cursorId")),
		Markups: buildSlice(container.Get("chartMarkups").Array(), func(item gjson.Result) models.ChartMarkup {
			return models.ChartMarkup{
				ID:        stringify(item.Get("id")),
				Name:      gStr(item.Get("name")),
				Data:      stringify(item.Get("data")),
				Frequency: gStr(item.Get("frequency")),
				Site:      gStr(item.Get("site")),
				CreatedAt: gStr(item.Get("createdAt")),
				UpdatedAt: gStr(item.Get("updatedAt")),
			}
		}),
	}, nil
}
