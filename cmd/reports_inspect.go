package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// ReportsInspectCmd retrieves the definition and filter criteria for a specific screen.
type ReportsInspectCmd struct {
	ScreenID string `arg:"" help:"Screen ID from 'reports list' output."`
	Coach    bool   `help:"Treat screen ID as a MarketSurge coach screen." default:"false"`
}

type screenInspect struct {
	ID          *string              `json:"id,omitempty"`
	Name        *string              `json:"name,omitempty"`
	Description *string              `json:"description,omitempty"`
	Type        *string              `json:"type,omitempty"`
	Filters     []screenFilter       `json:"filters"`
	FilterType  *string              `json:"filterType,omitempty"`
	ResultLimit *int                 `json:"resultLimit,omitempty"`
	SortBy      *screenSortBy        `json:"sortBy,omitempty"`
	LastResult  *screenResultSummary `json:"lastResult,omitempty"`
	CreatedAt   *string              `json:"createdAt,omitempty"`
	UpdatedAt   *string              `json:"updatedAt,omitempty"`
}

type screenFilter struct {
	Field   *string `json:"field,omitempty"`
	Operand *string `json:"operand,omitempty"`
	Value   *string `json:"value,omitempty"`
}

type screenSortBy struct {
	Field     *string `json:"field,omitempty"`
	Direction *string `json:"direction,omitempty"`
}

type screenResultSummary struct {
	Count       *int    `json:"count,omitempty"`
	Description *string `json:"description,omitempty"`
	UpdatedAt   *string `json:"updatedAt,omitempty"`
}

// Run executes the screen inspect query and writes a JSON array.
func (c *ReportsInspectCmd) Run(client *marketsurge.Client) error {
	return c.run(client, os.Stdout)
}

func (c *ReportsInspectCmd) run(client *marketsurge.Client, w io.Writer) error {
	req := marketsurge.NewScreenRequest(c.ScreenID)
	if c.Coach {
		coach := true
		req.CoachScreen = &coach
	}

	resp, err := client.Screen(context.Background(), req)
	if err != nil {
		return clientError("screen definition request failed", err)
	}

	if err := json.NewEncoder(w).Encode([]screenInspect{screenInspectFrom(resp)}); err != nil {
		return mserrors.NewAPIError("failed to write screen inspect output", err)
	}

	return nil
}

func screenInspectFrom(resp *marketsurge.ScreenResponse) screenInspect {
	if resp == nil || resp.User == nil || resp.User.Screen == nil {
		return screenInspect{Filters: []screenFilter{}}
	}

	screen := resp.User.Screen
	item := screenInspect{Filters: []screenFilter{}}
	item.ID = screen.ID
	item.Name = screen.Name
	item.Description = screen.Description
	item.Type = screen.Type
	if screen.FilterCriteria != nil {
		item.FilterType = screen.FilterCriteria.Type
	}
	if screen.ResultConfig != nil {
		item.ResultLimit = screen.ResultConfig.Limit
	}
	item.SortBy = screenSortByFrom(screen.ResultConfig)
	item.LastResult = screenResultSummaryFrom(screen.Result)
	item.CreatedAt = screen.CreatedAt
	item.UpdatedAt = screen.UpdatedAt
	item.Filters = screenFiltersFrom(screen.FilterCriteria)
	return item
}

func screenFiltersFrom(criteria *marketsurge.ScreenFilterCriteria) []screenFilter {
	if criteria == nil || len(criteria.Terms) == 0 {
		return []screenFilter{}
	}

	filters := make([]screenFilter, 0, len(criteria.Terms))
	for _, term := range criteria.Terms {
		filters = append(filters, screenFilter{
			Field:   screenFilterName(term.Left),
			Operand: term.Operand,
			Value:   screenFilterValue(term.Right),
		})
	}
	return filters
}

func screenSortByFrom(config *marketsurge.ScreenResultConfig) *screenSortBy {
	if config == nil || config.SortBy == nil {
		return nil
	}

	return &screenSortBy{
		Field:     config.SortBy.Field,
		Direction: config.SortBy.Direction,
	}
}

func screenResultSummaryFrom(result *marketsurge.ScreenResultSummary) *screenResultSummary {
	if result == nil {
		return nil
	}

	return &screenResultSummary{
		Count:       result.Count,
		Description: result.Description,
		UpdatedAt:   result.UpdatedAt,
	}
}

func screenFilterName(left *marketsurge.ScreenFilterTermLeft) *string {
	if left == nil {
		return nil
	}

	return left.Name
}

func screenFilterValue(right *marketsurge.ScreenFilterTermRight) *string {
	if right == nil {
		return nil
	}

	return right.Value
}
