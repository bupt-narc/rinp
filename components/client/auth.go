package client

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/bupt-narc/rinp/pkg/overlay"
)

var (
	defaultEmail    string = "example@example.com"
	defaultPassword string = "example@example.com"
)

var (
	baseURL string
	token   string
	vip     string
)

func login(email, password string) error {
	payload := url.Values{"email": {email}, "password": {password}}
	response, err := http.PostForm(baseURL+"/api/users/auth-via-email", payload)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// FIXME
	if response.StatusCode != 200 {
		return fmt.Errorf("login failed")
	}
	type LoginResponse struct {
		Token string                 `json:"token"`
		User  map[string]interface{} `json:"user"`
	}
	var loginResponse LoginResponse
	err = json.NewDecoder(response.Body).Decode(&loginResponse)
	if err != nil {
		return err
	}
	token = "User " + loginResponse.Token
	vip = loginResponse.User["profile"].(map[string]interface{})["vip"].(string)
	return nil
}

func setInfo(o *Option, email, password string) error {
	err := login(email, password)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("GET", baseURL+"/api/v1/rinp", nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", token)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", response.StatusCode)
	}

	type InfoResponse struct {
		ServerCIDR        []string `json:"server_cidr"`
		FirstProxyAddress string   `json:"first_proxy_address"`
		SchedulerAddress  string   `json:"scheduler_address"`
	}
	var infoResponse InfoResponse
	err = json.NewDecoder(response.Body).Decode(&infoResponse)
	if err != nil {
		return err
	}
	o.ServerCIDRs, err = overlay.StringToCIDRs(infoResponse.ServerCIDR)
	if err != nil {
		return err
	}
	o.ClientVirtualIP = net.ParseIP(vip)
	o.ProxyAddress = infoResponse.FirstProxyAddress
	// TODO schedulerAddress
	return nil
}

func setInfoByDefault(o *Option) error {
	baseURL = o.AuthBaseURL
	return setInfo(o, defaultEmail, defaultPassword)
}
