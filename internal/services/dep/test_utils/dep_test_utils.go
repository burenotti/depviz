package test_utils

import (
	"encoding/json"
)

type ServerResponse struct {
	Info Info `json:"info"`
}
type Info struct {
	Requires []string `json:"requires_dist"`
}

func NewServerResponse(deps ...string) []byte {
	r := ServerResponse{Info: Info{Requires: deps}}
	b, _ := json.Marshal(&r)
	return b
}
