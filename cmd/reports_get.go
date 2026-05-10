package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsGetCmd retrieves report data for a specific screen ID.
type ReportsGetCmd struct {
	ScreenID string   `arg:"" help:"Screen ID from 'reports list' output."`
	Columns  []string `help:"Response columns to include." default:"Symbol,CompanyName,ListRank,Price,PriceNetChg,PricePctChg,PricePctOff52WHigh,VolumePctChgVs50DAvgVolume,VolumeAvg50Day,MarketCapIntraday,CompositeRating,EPSRating,RSRating,AccDisRating,SMRRating,IndustryGroupRank,IndustryName,VolumeDollarAvg50D,IPODate,DowJonesKey,ChartingSymbol,DowJonesInstrumentType,DowJonesInstrumentSubType" env:"MARKETSURGE_AGENT_COLUMNS" sep:","`
}

// Run executes the report query and writes row objects as a JSON array.
func (c *ReportsGetCmd) Run(ctx context.Context, client *marketsurge.Client) error {
	return c.run(ctx, client, os.Stdout)
}

func (c *ReportsGetCmd) run(ctx context.Context, client *marketsurge.Client, w io.Writer) error {
	screenID := strings.TrimSpace(c.ScreenID)
	if screenID == "" {
		return mserrors.NewValidationError("screen ID is required", errors.New("empty screen ID"))
	}

	columns := make([]marketsurge.RunScreenResponseColumn, 0, len(c.Columns))
	for _, name := range c.Columns {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			return mserrors.NewValidationError("column names must not be empty", errors.New("empty column name"))
		}
		columns = append(columns, marketsurge.RunScreenResponseColumn{Name: marketsurge.ColumnName(trimmed)})
	}
	if len(columns) == 0 {
		return mserrors.NewValidationError("at least one column is required", errors.New("empty columns"))
	}

	resp, err := client.RunScreen(
		ctx,
		marketsurge.NewRunScreenRequest(screenID, columns),
	)
	if err != nil {
		if marketsurge.IsAuthError(err) {
			return mserrors.NewAuthenticationError("authentication failed", err)
		}
		if marketsurge.IsRateLimited(err) {
			return mserrors.NewHTTPError("rate limited", err, 429, "")
		}
		return mserrors.NewAPIError("API request failed", err)
	}

	if err := json.NewEncoder(w).Encode(reportsGetRows(resp)); err != nil {
		return mserrors.NewAPIError("failed to write report data", err)
	}

	return nil
}

func reportsGetRows(resp *marketsurge.RunScreenResponse) []map[string]any {
	if resp == nil || resp.User == nil || resp.User.RunScreen == nil {
		return []map[string]any{}
	}

	rows := make([]map[string]any, 0, len(resp.User.RunScreen.ResponseValues))
	for _, cells := range resp.User.RunScreen.ResponseValues {
		row := make(map[string]any, len(cells))
		for _, cell := range cells {
			if cell.MDItem != nil && cell.MDItem.Name != nil {
				row[*cell.MDItem.Name] = cell.Value
			}
		}
		rows = append(rows, row)
	}

	return rows
}
