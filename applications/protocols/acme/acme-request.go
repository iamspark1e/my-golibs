package acme

import (
	"errors"

	"github.com/go-resty/resty/v2"
)

type ACMERequestOption struct {
	Account ACMEAccount
	Method  string
	Header  map[string]string
	Payload interface{} // Protect & Signature parts will be generate when request
	Dirs    AcmeDirectory
}

// new-nonce don't need account info, always request (with HEAD), always response.
func AcmeNewNonce(nonceUrl string) (string, error) {
	client := resty.New()
	resp, err := client.R().Head(nonceUrl)
	if err != nil {
		return "", err
	}
	value := resp.Header().Get("replay-nonce")
	return value, nil
}
func ACMEGetRequest(url string, result interface{}) (*resty.Response, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/jose+json").
		ForceContentType("application/json").
		SetResult(&result).
		Get(url)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
func requiredACMEOptionCheck(option ACMERequestOption) error {
	if option.Dirs.NewNonce == "" {
		return errors.New("the ACME Directory is not initialized")
	}
	return nil
}

// PostAsGet: option.Payload = ""
func ACMEPostRequest(url string, option ACMERequestOption, result interface{}) (*resty.Response, error) {
	err := requiredACMEOptionCheck(option)
	if err != nil {
		return nil, err
	}
	client := resty.New()
	nonce, err := AcmeNewNonce(option.Dirs.NewNonce)
	if err != nil {
		return nil, err
	}
	reqBody, err := JwsEncodeJSON(option.Payload, option.Account.PrivKey, nonce, url)
	if err != nil {
		return nil, err
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/jose+json").
		ForceContentType("application/json").
		SetBody(reqBody).
		SetResult(&result).
		Post(url)
	if err != nil {
		return resp, err
	}
	return resp, nil
}
