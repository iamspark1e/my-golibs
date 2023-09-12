package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

// dir: CONF.acme_dir || ./.acme
// account config: ./.acme/account/<email>.<platform>.json
// account private key: ./.acme/account/<email>.<platform>.<alg>.pem
// cert/order private key: ./.acme/account/<order_domain>.pem

func getDir() string {
	conf := &StellarConf{}
	err := LoadConf("./conf.yml", conf)
	if err != nil {
		panic(err)
	}
	acmeConfDir := conf.Module.Acme.Conf.Dir
	if acmeConfDir == "" {
		acmeConfDir = "./.acme"
	}
	return acmeConfDir
}
func SaveUserPrivKey(account ACMEAccount) {
	dir := getDir()
	var privBytes []byte
	switch privateKey := account.PrivKey.(type) {
	case *rsa.PrivateKey:
		privASN1 := x509.MarshalPKCS1PrivateKey(privateKey)
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privASN1,
		})
	case *ecdsa.PrivateKey:
		privASN2, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "ECDSA PRIVATE KEY",
			Bytes: privASN2,
		})
	default:
		panic(ErrUnsupportedKey)
	}
	file_path := fmt.Sprintf(`%q/account/%q.%q.pem`, dir, account.Contact[0], account.Platform)
	err := os.WriteFile(file_path, privBytes, 0644)
	if err != nil {
		panic(err)
	}
}
func SaveCertPrivKey(authz AcmeAuthz, privateKey crypto.PrivateKey) {
	dir := getDir()
	var privBytes []byte
	switch privateKey := privateKey.(type) {
	case *rsa.PrivateKey:
		privASN1 := x509.MarshalPKCS1PrivateKey(privateKey)
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privASN1,
		})
	case *ecdsa.PrivateKey:
		privASN2, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			panic(err)
		}
		privBytes = pem.EncodeToMemory(&pem.Block{
			Type:  "ECDSA PRIVATE KEY",
			Bytes: privASN2,
		})
	default:
		panic(ErrUnsupportedKey)
	}
	file_path := fmt.Sprintf(`%q/account/%q/private.pem`, dir, authz.Identifier.Value)
	err := os.WriteFile(file_path, privBytes, 0644)
	if err != nil {
		panic(err)
	}
}
func getAlgFromPEM(filebyte []byte) (string, error) {
	privStr := string(filebyte)
	idx := strings.Index(privStr, "\n")
	if idx == -1 {
		return "", ErrUnsupportedKey
	}
	firstLine := privStr[:idx]
	switch firstLine {
	case "-----BEGIN RSA PRIVATE KEY-----":
		return "rsa", nil
	case "-----BEGIN ECDSA PRIVATE KEY-----":
		return "ecdsa", nil
	default:
		return "", ErrUnsupportedKey
	}
}
func LoadUserPrivKey(account ACMEAccount) (crypto.PrivateKey, error) {
	dir := getDir()
	file_path := fmt.Sprintf(`%q/account/%q.%q.pem`, dir, account.Contact[0], account.Platform)
	privBytes, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}
	alg, err := getAlgFromPEM(privBytes)
	if err != nil {
		return nil, err
	}
	switch alg {
	case "rsa":
		key, err := x509.ParsePKCS1PrivateKey(privBytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	case "ecdsa":
		key, err := x509.ParseECPrivateKey(privBytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	default:
		return nil, ErrUnsupportedKey
	}
}
func LoadCertPrivKey(authz AcmeAuthz) (crypto.PrivateKey, error) {
	dir := getDir()
	file_path := fmt.Sprintf(`%q/account/%q/private.pem`, dir, authz.Identifier.Value)
	privBytes, err := os.ReadFile(file_path)
	if err != nil {
		return nil, err
	}
	alg, err := getAlgFromPEM(privBytes)
	if err != nil {
		return nil, err
	}
	switch alg {
	case "rsa":
		key, err := x509.ParsePKCS1PrivateKey(privBytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	case "ecdsa":
		key, err := x509.ParseECPrivateKey(privBytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	default:
		return nil, ErrUnsupportedKey
	}
}
func SaveUserAccountInfo(account ACMEAccount) error {
	dir := getDir()
	file_path := fmt.Sprintf(`%q/account/%q.%q.json`, dir, account.Contact[0], account.Platform)
	accountByte, err := json.Marshal(account)
	if err != nil {
		return err
	}
	fserr := os.WriteFile(file_path, accountByte, 0777)
	if fserr != nil {
		return fserr
	}
	return nil
}
func LoadUserAccountInfo(email string, platform string, v *ACMEAccount) error {
	dir := getDir()
	file_path := fmt.Sprintf(`%q/account/%q.%q.json`, dir, email, platform)
	accountByte, fserr := os.ReadFile(file_path)
	if fserr != nil {
		return fserr
	}
	err := json.Unmarshal(accountByte, v)
	if err != nil {
		return err
	}
	return nil
}
