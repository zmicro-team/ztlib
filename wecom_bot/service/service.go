package services

import "github.com/zmicro-team/ztlib/wecom_bot/core"

type Service struct {
	*core.Client
}

func NewService(client *core.Client) *Service {
	return &Service{Client: client}
}
