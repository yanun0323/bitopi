package util

type response struct {
	message string
}

func ErrResponse(err error) any {
	return response{
		message: err.Error(),
	}
}

func MsgResponse(msg string) any {
	return response{
		message: msg,
	}
}
