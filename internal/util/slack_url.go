package util

type Url string

const (
	PostChat    Url = "https://slack.com/api/chat.postMessage"
	PostView    Url = "https://slack.com/api/views.open"
	PostCommand Url = ""
)

func NewUrl(url string) Url {
	return Url(url)
}

func (u Url) String() string {
	return string(u)
}
