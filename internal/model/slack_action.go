package model

type SlackMessageButton struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Style string `json:"style"` /* default, primary, danger */
	Type  string `json:"type"`
	Value string `json:"value"`
}

/*
style: default, primary, danger
*/
func NewSlackMessageActionButton(style, value, text string) SlackMessageButton {
	return SlackMessageButton{
		Name:  "direct_msg_action",
		Text:  text,
		Style: style,
		Type:  "button",
		Value: value,
	}
}

type SlackActionPayload struct {
	Payload map[string]interface{} `json:"payload"`
}
