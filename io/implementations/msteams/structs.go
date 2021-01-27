package msteams

import "time"

type Outgoing struct {
	Type           string    `json:"type"`
	ID             string    `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	LocalTimestamp string    `json:"localTimestamp"`
	ServiceURL     string    `json:"serviceUrl"`
	ChannelID      string    `json:"channelId"`
	From           struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AadObjectID string `json:"aadObjectId"`
	} `json:"from"`
	Conversation struct {
		IsGroup          bool        `json:"isGroup"`
		ID               string      `json:"id"`
		Name             interface{} `json:"name"`
		ConversationType string      `json:"conversationType"`
		TenantID         string      `json:"tenantId"`
	} `json:"conversation"`
	Recipient        interface{}   `json:"recipient"`
	TextFormat       string        `json:"textFormat"`
	AttachmentLayout interface{}   `json:"attachmentLayout"`
	MembersAdded     []interface{} `json:"membersAdded"`
	MembersRemoved   []interface{} `json:"membersRemoved"`
	TopicName        interface{}   `json:"topicName"`
	HistoryDisclosed interface{}   `json:"historyDisclosed"`
	Locale           string        `json:"locale"`
	Text             string        `json:"text"`
	Speak            interface{}   `json:"speak"`
	InputHint        interface{}   `json:"inputHint"`
	Summary          interface{}   `json:"summary"`
	SuggestedActions interface{}   `json:"suggestedActions"`
	Attachments      []struct {
		ContentType  string      `json:"contentType"`
		ContentURL   interface{} `json:"contentUrl"`
		Content      string      `json:"content"`
		Name         interface{} `json:"name"`
		ThumbnailURL interface{} `json:"thumbnailUrl"`
	} `json:"attachments"`
	Entities []struct {
		Type     string `json:"type"`
		Locale   string `json:"locale"`
		Country  string `json:"country"`
		Platform string `json:"platform"`
		Timezone string `json:"timezone"`
	} `json:"entities"`
	ChannelData struct {
		TeamsChannelID string `json:"teamsChannelId"`
		TeamsTeamID    string `json:"teamsTeamId"`
		Channel        struct {
			ID string `json:"id"`
		} `json:"channel"`
		Team struct {
			ID string `json:"id"`
		} `json:"team"`
		Tenant struct {
			ID string `json:"id"`
		} `json:"tenant"`
	} `json:"channelData"`
	Action        interface{} `json:"action"`
	ReplyToID     interface{} `json:"replyToId"`
	Value         interface{} `json:"value"`
	Name          interface{} `json:"name"`
	RelatesTo     interface{} `json:"relatesTo"`
	Code          interface{} `json:"code"`
	LocalTimezone string      `json:"localTimezone"`
}
