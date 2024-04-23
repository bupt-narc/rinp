package client

import (
	"encoding/json"
	"fmt"
	"github.com/bupt-narc/rinp/pkg/overlay"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var (
	defaultEmail    string = "example@example.com"
	defaultPassword string = "example@example.com"
)

var (
	baseURL string
)

func setInfo(o *Option, email, password string) error {
	data := url.Values{}
	data.Set("identity", email)
	data.Set("password", password)
	// Body parameters could be sent as multipart/form-data and tell out Content-Type.
	request, err := http.NewRequest("POST", baseURL+"/api/collections/users/auth-with-password", strings.NewReader(data.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	type InfoResponse struct {
		Record struct {
			ServerCIDR        string `json:"serverCIDR"`
			FirstProxyAddress string `json:"firstProxyAddress"`
			SchedulerAddress  string `json:"schedulerAddress"`
			VIP               string `json:"vip"`
		} `json:"record"`
	}
	var infoResponse InfoResponse
	err = json.NewDecoder(response.Body).Decode(&infoResponse)
	if err != nil {
		return err
	}
	o.ServerCIDRs, err = overlay.StringToCIDRs(strings.Split(infoResponse.Record.ServerCIDR, ","))
	if err != nil {
		return err
	}
	o.ClientVirtualIP = net.ParseIP(infoResponse.Record.VIP)
	if o.ClientVirtualIP == nil {
		return fmt.Errorf("invalid vip: %s", infoResponse.Record.VIP)
	}
	o.ProxyAddress = infoResponse.Record.FirstProxyAddress
	o.SchedulerAddress = infoResponse.Record.SchedulerAddress
	return nil
}

func setInfoByDefault(o *Option) error {
	baseURL = o.AuthBaseURL
	return setInfo(o, defaultEmail, defaultPassword)
}
