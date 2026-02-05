package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type GridClient struct {
	BaseURL string
}

func (g *GridClient) CallVerify(uuid string) ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/internal/identity/verify?uuid=%s", g.BaseURL, uuid))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}
