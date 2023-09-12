// Specially thanks to [Open Source Code](https://github.com/HewlettPackard/docker-machine-oneview/blob/master/vendor/golang.org/x/crypto/acme/internal/acme/jws.go)
package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
)

// ErrUnsupportedKey is returned when an unsupported key type is encountered.
var ErrUnsupportedKey = errors.New("acme: unknown key type; only RSA and ECDSA are supported")

// jwkEncode encodes public part of an RSA or ECDSA key into a JWK.
// The result is also suitable for creating a JWK thumbprint.
// https://tools.ietf.org/html/rfc7517
func jwkEncode(pub crypto.PublicKey) (string, error) {
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		// https://tools.ietf.org/html/rfc7518#section-6.3.1
		n := pub.N
		e := big.NewInt(int64(pub.E))
		// Field order is important.
		// See https://tools.ietf.org/html/rfc7638#section-3.3 for details.
		return fmt.Sprintf(`{"e":"%s","kty":"RSA","n":"%s"}`,
			base64.RawURLEncoding.EncodeToString(e.Bytes()),
			base64.RawURLEncoding.EncodeToString(n.Bytes()),
		), nil
	case *ecdsa.PublicKey:
		// https://tools.ietf.org/html/rfc7518#section-6.2.1
		p := pub.Curve.Params()
		n := p.BitSize / 8
		if p.BitSize%8 != 0 {
			n++
		}
		x := pub.X.Bytes()
		if n > len(x) {
			x = append(make([]byte, n-len(x)), x...)
		}
		y := pub.Y.Bytes()
		if n > len(y) {
			y = append(make([]byte, n-len(y)), y...)
		}
		// Field order is important.
		// See https://tools.ietf.org/html/rfc7638#section-3.3 for details.
		return fmt.Sprintf(`{"crv":"%s","kty":"EC","x":"%s","y":"%s"}`,
			p.Name,
			base64.RawURLEncoding.EncodeToString(x),
			base64.RawURLEncoding.EncodeToString(y),
		), nil
	}
	return "", ErrUnsupportedKey
}

// FIXME: support any legel length of crypto?
func GetKeyAlgorithm(pub crypto.PublicKey) (string, error) {
	switch pub.(type) {
	case *rsa.PublicKey:
		alg := "RS256"
		return alg, nil
	case *ecdsa.PublicKey:
		alg := "HS256"
		return alg, nil
	default:
		return "", ErrUnsupportedKey
	}
}

// JWKThumbprint creates a JWK thumbprint out of pub
// as specified in https://tools.ietf.org/html/rfc7638.
func JWKThumbprint(pub crypto.PublicKey) (string, error) {
	jwk, err := jwkEncode(pub)
	if err != nil {
		return "", err
	}
	b := sha256.Sum256([]byte(jwk))
	return base64.RawURLEncoding.EncodeToString(b[:]), nil
}

// JwsEncodeJson helpers
func jwsEcdsaEncodeJSON(claimset interface{}, key *ecdsa.PrivateKey, nonce string, url string) ([]byte, error) {
	jwk, err := jwkEncode(key.Public())
	if err != nil {
		return nil, err
	}
	phead := fmt.Sprintf(`{"alg":"HS256","jwk":%s,"nonce":%q,"url":%q}`, jwk, nonce, url)
	phead = base64.RawURLEncoding.EncodeToString([]byte(phead))
	cs, err := json.Marshal(claimset)
	if err != nil {
		return nil, err
	}
	payload := base64.RawURLEncoding.EncodeToString(cs)
	h := sha256.New()
	h.Write([]byte(phead + "." + payload))
	sig, err := key.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return nil, err
	}
	enc := struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Sig       string `json:"signature"`
	}{
		Protected: phead,
		Payload:   payload,
		Sig:       base64.RawURLEncoding.EncodeToString(sig),
	}
	return json.Marshal(&enc)
}
func jwsRsaEncodeJSON(claimset interface{}, key *rsa.PrivateKey, nonce string, url string) ([]byte, error) {
	jwk, err := jwkEncode(key.Public())
	if err != nil {
		return nil, err
	}
	phead := fmt.Sprintf(`{"alg":"RS256","jwk":%s,"nonce":%q,"url":%q}`, jwk, nonce, url)
	phead = base64.RawURLEncoding.EncodeToString([]byte(phead))
	cs, err := json.Marshal(claimset)
	if err != nil {
		return nil, err
	}
	payload := base64.RawURLEncoding.EncodeToString(cs)
	h := sha256.New()
	h.Write([]byte(phead + "." + payload))
	sig, err := key.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return nil, err
	}
	enc := struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Sig       string `json:"signature"`
	}{
		Protected: phead,
		Payload:   payload,
		Sig:       base64.RawURLEncoding.EncodeToString(sig),
	}
	return json.Marshal(&enc)
}

// jwsEncodeJSON signs claimset using provided key and a nonce.
// The result is serialized in JSON format.
// See https://tools.ietf.org/html/rfc7515#section-7.
func JwsEncodeJSON(claimset interface{}, privKey crypto.PrivateKey, nonce string, url string) ([]byte, error) {
	switch key := privKey.(type) {
	case *rsa.PrivateKey:
		data, err := jwsRsaEncodeJSON(claimset, key, nonce, url)
		return data, err
	case *ecdsa.PrivateKey:
		data, err := jwsEcdsaEncodeJSON(claimset, key, nonce, url)
		return data, err
	default:
		return nil, ErrUnsupportedKey
	}
}

func JwsEncodeJSONWithKid(claimset interface{}, key crypto.Signer, nonce string, url string, kid string) ([]byte, error) {
	alg, err := GetKeyAlgorithm(key.Public())
	if err != nil {
		return nil, err
	}
	phead := fmt.Sprintf(`{"alg":%q,"kid":%q,"nonce":%q,"url":%q}`, alg, kid, nonce, url)
	phead = base64.RawURLEncoding.EncodeToString([]byte(phead))
	var payload string
	if claimset == nil {
		claimset = ""
		payload = ""
	} else {
		cs, err := json.Marshal(claimset)
		if err != nil {
			return nil, err
		}
		payload = base64.RawURLEncoding.EncodeToString(cs)
	}
	h := sha256.New()
	h.Write([]byte(phead + "." + payload))
	sig, err := key.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return nil, err
	}
	enc := struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Sig       string `json:"signature"`
	}{
		Protected: phead,
		Payload:   payload,
		Sig:       base64.RawURLEncoding.EncodeToString(sig),
	}
	return json.Marshal(&enc)
}

func JwsEncodeStringWithKid(payload string, key crypto.Signer, nonce string, url string, kid string) ([]byte, error) {
	alg, err := GetKeyAlgorithm(key.Public())
	if err != nil {
		return nil, err
	}
	phead := fmt.Sprintf(`{"alg":%q,"kid":%q,"nonce":%q,"url":%q}`, alg, kid, nonce, url)
	phead = base64.RawURLEncoding.EncodeToString([]byte(phead))
	payload = base64.RawURLEncoding.EncodeToString([]byte(payload))
	h := sha256.New()
	h.Write([]byte(phead + "." + payload))
	sig, err := key.Sign(rand.Reader, h.Sum(nil), crypto.SHA256)
	if err != nil {
		return nil, err
	}
	enc := struct {
		Protected string `json:"protected"`
		Payload   string `json:"payload"`
		Sig       string `json:"signature"`
	}{
		Protected: phead,
		Payload:   payload,
		Sig:       base64.RawURLEncoding.EncodeToString(sig),
	}
	return json.Marshal(&enc)
}
