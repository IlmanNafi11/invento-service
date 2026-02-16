package supabase

import (
	"context"
	"invento-service/config"

	"github.com/supabase-community/supabase-go"
)

type Client struct {
	*supabase.Client
	URL    string
	Config *config.SupabaseConfig
}

var globalClient *Client

func Init(cfg *config.SupabaseConfig) (*Client, error) {
	client, err := supabase.NewClient(cfg.URL, cfg.ServiceKey, &supabase.ClientOptions{
		Schema: "public",
	})
	if err != nil {
		return nil, err
	}

	globalClient = &Client{
		Client: client,
		URL:    cfg.URL,
		Config: cfg,
	}

	return globalClient, nil
}

func Get() *Client {
	return globalClient
}

func (c *Client) Ping(ctx context.Context) error {
	_, _, err := c.From("user_profiles").Select("*", "", false).Execute()
	return err
}
