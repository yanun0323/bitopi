package util

type Url string

const (
	PostChat Url = "https://slack.com/api/chat.postMessage"
	PostView Url = "https://slack.com/api/views.open"
	PostHome Url = "https://slack.com/api/views.publish"

	GetChat      Url = "https://slack.com/api/conversations.replies"
	GetPermalink Url = "https://slack.com/api/chat.getPermalink"
)

func (u Url) String() string {
	return string(u)
}
