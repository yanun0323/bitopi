package util

import (
	"io"
	"net/http"
)

type HttpRequestOption struct {
	Method       string
	Url          string
	Token        string
	IsUrlencoded bool
}

func HttpRequest(opt HttpRequestOption) ([]byte, int, error) {
	req, err := http.NewRequest(opt.Method, opt.Url, nil)
	if err != nil {
		return nil, 0, err
	}

	if opt.IsUrlencoded {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	} else {
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	}

	if len(opt.Token) != 0 {
		req.Header.Set("Authorization", "Bearer "+opt.Token) //"Bearer "+viper.GetString("token.maid")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	return buf, res.StatusCode, nil
}
