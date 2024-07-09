package model

type SlackVerificationResponse struct {
	Challenge string `json:"challenge"`
}

type SlackTypeCheck struct {
	Type      string `json:"type"`
	Challenge string `json:"challenge"`
}

type SlackVerification struct {
	Challenge string `json:"challenge"`
}
