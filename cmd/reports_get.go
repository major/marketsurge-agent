package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsGetCmd retrieves report data for a specific screen ID.
type ReportsGetCmd struct {
	ScreenID string   `arg:"" help:"Screen ID from 'reports list' output."`
	Columns  []string `help:"Response columns to include." default:"Symbol,CompanyName,ListRank,Price,PriceNetChg,PricePctChg,PricePctOff52WHigh,VolumePctChgVs50DAvgVolume,VolumeAvg50Day,MarketCapIntraday,CompositeRating,EPSRating,RSRating,AccDisRating,SMRRating,IndustryGroupRank,IndustryName,VolumeDollarAvg50D,IPODate,DowJonesKey,ChartingSymbol,DowJonesInstrumentType,DowJonesInstrumentSubType" env:"MARKETSURGE_AGENT_COLUMNS" sep:","`
}

// Run executes the report query and writes row objects as a JSON array.
func (c *ReportsGetCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *ReportsGetCmd) run(client *marketsurge.Client, w io.Writer) error {
	resp, err := client.RunScreen(
		context.Background(),
		marketsurge.NewRunScreenRequest(c.ScreenID, c.Columns),
	)
	if err != nil {
		if marketsurge.IsAuthError(err) {
			return mserrors.NewAuthenticationError("authentication failed", err)
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
