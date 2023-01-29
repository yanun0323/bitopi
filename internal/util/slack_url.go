package util

type Url string

const (
	PostChat     Url = "https://slack.com/api/chat.postMessage"
	PostView     Url = "https://slack.com/api/views.open"
	GetPermalink Url = "https://slack.com/api/chat.getPermalink"
	GetThread    Url = "https://slack.com/api/conversations.replies"
)

func (u Url) String() string {
	return string(u)
}
