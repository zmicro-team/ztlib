package message_send

import "github.com/zmicro-team/ztlib/wecom_bot/consts"

type TextRequest struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

type TextResponse struct {
	consts.CommonResp `json:",inline"`
}
