package variable

import (
	"encoding/json"
	"meta/meta-feishu/oapi/card"
)

type RestartServer struct {
	Id           *string              `json:"id,omitempty"`
	Name         *string              `json:"name,omitempty"`
	NameShort    *string              `json:"name_short,omitempty"`
	Type         *string              `json:"type,omitempty"`
	Handler      *string              `json:"handler,omitempty"`
	Reason       *string              `json:"reason,omitempty"`
	Source       *string              `json:"source,omitempty"`
	ServerLog    *string              `json:"server_log,omitempty"`
	Countdown    *string              `json:"countdown,omitempty"`
	Blocker      *string              `json:"blocker,omitempty"`
	CardLink     *card.MessageCardURL `json:"card_link,omitempty"`
	VariableJson *string              `json:"variable_json,omitempty"`
}

type RestartServerDataBuilder struct {
	server RestartServer
}

func NewRestartServerDataBuilder() *RestartServerDataBuilder {
	return &RestartServerDataBuilder{
		server: RestartServer{},
	}
}

func (b *RestartServerDataBuilder) Id(id string) *RestartServerDataBuilder {
	b.server.Id = &id
	return b
}

func (b *RestartServerDataBuilder) Name(name *string) *RestartServerDataBuilder {
	b.server.Name = name
	return b
}

func (b *RestartServerDataBuilder) NameShort(nameShort *string) *RestartServerDataBuilder {
	b.server.NameShort = nameShort
	return b
}

func (b *RestartServerDataBuilder) Type(t string) *RestartServerDataBuilder {
	b.server.Type = &t
	return b
}

func (b *RestartServerDataBuilder) Handler(handler string) *RestartServerDataBuilder {
	b.server.Handler = &handler
	return b
}

func (b *RestartServerDataBuilder) Reason(reason *string) *RestartServerDataBuilder {
	b.server.Reason = reason
	return b
}

func (b *RestartServerDataBuilder) Source(source string) *RestartServerDataBuilder {
	b.server.Source = &source
	return b
}

func (b *RestartServerDataBuilder) ServerLog(serverLog string) *RestartServerDataBuilder {
	b.server.ServerLog = &serverLog
	return b
}

func (b *RestartServerDataBuilder) Countdown(countdown string) *RestartServerDataBuilder {
	b.server.Countdown = &countdown
	return b
}

func (b *RestartServerDataBuilder) Blocker(blocker string) *RestartServerDataBuilder {
	b.server.Blocker = &blocker
	return b
}

func (b *RestartServerDataBuilder) CardLink(cardLink *card.MessageCardURL) *RestartServerDataBuilder {
	b.server.CardLink = cardLink
	return b
}

func (b *RestartServerDataBuilder) Build() *RestartServer {
	jsonData, err := json.Marshal(b.server)
	if err == nil {
		jsonString := string(jsonData)
		b.server.VariableJson = &jsonString
	}
	return &b.server
}
