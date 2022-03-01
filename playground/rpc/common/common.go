package common

import (
	"errors"
	"fmt"
)

type Request struct {
	Id string
}

type Response struct {
	Value string
}

type Handler struct{}

func (h *Handler) Execute(req Request, res *Response) error {
	if req.Id == "" {
		return errors.New("id must be specified")
	}
	res.Value = fmt.Sprintf("Value: %s", req.Id)
	return nil
}
