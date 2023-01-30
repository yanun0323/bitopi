package model

type SlackActionButton struct {
	Name  string `json:"name"`
	Text  string `json:"text"`
	Style string `json:"style"` /* default, primary, danger */
	Type  string `json:"type"`
	Value string `json:"value"`
}

/*
style: default, primary, danger
*/
func NewSlackActionButton(style, value, text string) SlackActionButton {
	return SlackActionButton{
		Name:  "direct_msg_action",
		Text:  text,
		Style: style,
		Type:  "button",
		Value: value,
	}
}

type SlackInteractionPayload struct {
	Payload map[string]interface{} `json:"payload"`
}
