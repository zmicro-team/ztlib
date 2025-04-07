package core

import (
	"github.com/go-resty/resty/v2"
	"github.com/zmicro-team/ztlib/wecom_bot/consts"
)

// bot

type Config struct {
	Key         string
	IsEnableLog bool
}

type Client struct {
	http   *resty.Client
	config Config
	log    Log
}

func NewClient(cfg Config) *Client {
	c := &Client{
		http:   resty.New(),
		config: cfg,
		log: &defaultLog{
			AppName: "wecombot",
		},
	}
	c.http = c.http.SetBaseURL(consts.BASE_URL)
	return c
}

// send text
