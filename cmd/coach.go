package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/major/marketsurge-go/marketsurge"

	mserrors "github.com/major/marketsurge-agent/internal/errors"
)

// CoachCmd discovers MarketSurge curated watchlists and screens.
type CoachCmd struct {
	Type string `help:"Filter by content type: watchlist, screen, or all." default:"all" enum:"all,watchlist,screen"`
}

// Run executes the coach tree query and writes a flat JSON array.
func (c *CoachCmd) Run(ctx context.Context, client *marketsurge.Client) error {
	return c.run(ctx, client, os.Stdout)
}

func (c *CoachCmd) run(ctx context.Context, client *marketsurge.Client, w io.Writer) error {
	resp, err := client.CoachTree(ctx, marketsurge.NewCoachTreeRequest())
	if err != nil {
		return clientError("coach tree request failed", err)
	}

	nodes := coachNodesFrom(resp, c.Type)
	if err := json.NewEncoder(w).Encode(nodes); err != nil {
		return mserrors.NewAPIError("failed to write coach output", err)
	}

	return nil
}

type coachNode struct {
	ID          *string          `json:"id,omitempty"`
	Name        *string          `json:"name,omitempty"`
	Type        *string          `json:"type,omitempty"`
	ContentType *string          `json:"contentType,omitempty"`
	Category    string           `json:"category"`
	Children    []coachChildNode `json:"children,omitempty"`
	URL         *string          `json:"url,omitempty"`
	ReferenceID *string          `json:"referenceId,omitempty"`
}

type coachChildNode struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

func coachNodesFrom(resp *marketsurge.CoachTreeResponse, filterType string) []coachNode {
	if resp == nil || resp.User == nil {
		return []coachNode{}
	}

	if filterType == "" {
		filterType = "all"
	}

	nodes := make([]coachNode, 0, len(resp.User.Watchlists)+len(resp.User.Screens))
	if filterType == "all" || filterType == "watchlist" {
		nodes = append(nodes, coachNodesFromTreeNodes(resp.User.Watchlists, "watchlist")...)
	}
	if filterType == "all" || filterType == "screen" {
		nodes = append(nodes, coachNodesFromTreeNodes(resp.User.Screens, "screen")...)
	}

	if len(nodes) == 0 {
		return []coachNode{}
	}

	return nodes
}

func coachNodesFromTreeNodes(nodes []marketsurge.CoachTreeNode, category string) []coachNode {
	if len(nodes) == 0 {
		return []coachNode{}
	}

	result := make([]coachNode, 0, len(nodes))
	for i := range nodes {
		node := nodes[i]
		item := coachNode{
			ID:          node.ID,
			Name:        node.Name,
			Type:        node.Type,
			ContentType: node.ContentType,
			Category:    category,
			URL:         node.URL,
			ReferenceID: node.ReferenceID,
		}
		if len(node.Children) > 0 {
			item.Children = make([]coachChildNode, 0, len(node.Children))
			for j := range node.Children {
				child := node.Children[j]
				item.Children = append(item.Children, coachChildNode{
					ID:   child.ID,
					Name: child.Name,
					Type: child.Type,
				})
			}
		}
		result = append(result, item)
	}

	return result
}
