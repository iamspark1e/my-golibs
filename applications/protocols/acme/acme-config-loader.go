package acme

import (
	"os"

	"crypto"

	"gopkg.in/yaml.v3"
)

// structors for program
type ACMEExternalBinding struct {
	Kid     string `json:"eab_kid"`
	HmacKey string `json:"eab_hmac_key"`
}
type ACMEAccount struct {
	AccountUrl      string
	ExternalBinding ACMEExternalBinding
	PrivKey         crypto.PrivateKey
	Kid             string
	Contact         []string
	Platform        string
}
type ACMEInstance struct {
	Directory    AcmeDirectory
	Order        AcmeNewOrder
	OrderPrivKey crypto.PrivateKey
}

// structors for program end

type AcmeDirectory struct {
	NewNonce   string `json:"newNonce"`
	NewAccount string `json:"newAccount"`
	NewOrder   string `json:"newOrder"`
	RevokeCert string `json:"revokeCert"`
	KeyChange  string `json:"keyChange"`
	Meta       struct {
		TermsOfService          string   `json:"termsOfService"`
		Website                 string   `json:"website"`
		CaaIdentities           []string `json:"caaIdentities"`
		ExternalAccountRequired bool     `json:"externalAccountRequired"`
	} `json:"meta"`
}

type AcmeNewAccountPayload struct {
	TermsOfServiceAgreed bool     `json:"termsOfServiceAgreed"`
	Contact              []string `json:"contact"`
	// externalAccountBinding ExternalAccountBinding
}

type AcmeNewAccount struct {
	Key struct {
		Kty string `json:"kty"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"key"`
	Contact   []string `json:"contact"`
	InitialIp string   `json:"initialIp"`
	CreatedAt string   `json:"createdAt"`
	Status    string   `json:"status"`
}

type AcmeOrderIdentifier struct {
	Type  string `json:"type"` // `dns` or `ipv4`
	Value string `json:"value"`
}

type AcmeNewOrderPayload struct {
	Identifiers []AcmeOrderIdentifier `json:"identifiers"`
}
type AcmeNewOrder struct {
	Status         string                `json:"status"`
	Expires        string                `json:"expires"`
	Identifiers    []AcmeNewOrderPayload `json:"identifiers"`
	Authorizations []string              `json:"authorizations"`
	Finalize       string                `json:"finalize"`
}
type AcmeChallenge struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Url    string `json:"url"`
	Token  string `json:"token"`
}
type AcmeAuthz struct {
	Identifier AcmeOrderIdentifier `json:"identifier"`
	Status     string              `json:"status"`
	Expires    string              `json:"expires"`
	Challenges []AcmeChallenge     `json:"challenges"`
	Wildcard   bool                `json:"wildcard"`
}
type AcmeChall struct {
	Type             string `json:"type"`
	Status           string `json:"status"`
	Url              string `json:"url"`
	Token            string `json:"token"`
	Validated        string `json:"validated"`
	ValidationRecord struct {
		Url             string   `json:"url"`
		Hostname        string   `json:"Hostname"`
		Port            string   `json:"Port"`
		AddressUsed     string   `json:"addressUsed"`
		AddressResolved []string `json:"addressResolved"`
	} `json:"validationRecord"`
}

type AcmeFinalizeRes struct {
	Status         string                `json:"status"`
	Expires        string                `json:"expires"`
	Identifiers    []AcmeOrderIdentifier `json:"identifiers"`
	Authorizations []string              `json:"authorizations"`
	Finalize       string                `json:"finalize"`
	Certificate    string                `json:"certificate"`
}

type StellarConf struct {
	Mode              string             `yaml:"mode"`
	Port              int                `yaml:"port"`
	LogLevel          string             `yaml:"log_level"`
	Feature           StellarConfFeature `yaml:"feature"`
	Module            StellarConfModule  `yaml:"module"`
	ServicePrivKeyDir string             `yaml:"service_priv_key_dir"`
	Email             string             `yaml:"email"`
	HMACSeed          string             `yaml:"hmac_seed"`
}
type StellarConfFeature struct {
	Webui       string `yaml:"web_ui"`
	WebuiSpa    bool   `yaml:"web_ui_spa"`
	DisableGrpc bool   `yaml:"disable_grpc"`
}
type StellarConfModule struct {
	Acme struct {
		Enable bool              `yaml:"enable"`
		Conf   StellarModuleAcme `yaml:"config"`
	} `yaml:"acme"`
	Dns struct {
		Enable bool             `yaml:"enable"`
		Conf   StellarModuleDNS `yaml:"config"`
	} `yaml:"dns"`
}
type StellarModuleAcme struct {
	Dir                    string                `yaml:"dir"`
	ExternalAccountBinding []ACMEExternalBinding `yaml:"external_account_binding"`
	ExpireCheckDuration    int                   `yaml:"expire_check_duration"`
}
type StellarModuleDNS struct {
	Provider      string      `yaml:"provider"`
	Authorization interface{} `yaml:"auth"`
}

func LoadConf(filepath string, cnf interface{}) error {
	yamlFile, err := os.ReadFile(filepath)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, cnf)
	}
	return err
}
