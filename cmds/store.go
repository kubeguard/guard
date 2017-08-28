package cmds

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/homedir"
)

type CertStore struct {
	dir    string
	Expiry time.Duration
}

func NewCertStore() (*CertStore, error) {
	dir := filepath.Join(homedir.HomeDir(), ".guard", "pki")
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, fmt.Errorf("Failed to create dir `%s`. Reason: %v.", dir, err)
	}
	return &CertStore{dir: dir}, nil
}

func (s *CertStore) Location() string {
	return s.dir
}

func (s *CertStore) Filename(cfg cert.Config) string {
	if len(cfg.Organization) == 0 {
		return cfg.CommonName
	}
	return cfg.CommonName + "@" + cfg.Organization[0]
}

func (s *CertStore) IsExists(name string) bool {
	if _, err := os.Stat(s.CertFile(name)); err == nil {
		return true
	}
	if _, err := os.Stat(s.KeyFile(name)); err == nil {
		return true
	}
	return false
}

func (s *CertStore) PairExists(name string) bool {
	if _, err := os.Stat(s.CertFile(name)); err == nil {
		if _, err := os.Stat(s.KeyFile(name)); err == nil {
			return true
		}
	}
	return false
}

func (s *CertStore) CertFile(name string) string {
	return filepath.Join(s.dir, strings.ToLower(name)+".crt")
}

func (s *CertStore) KeyFile(name string) string {
	return filepath.Join(s.dir, strings.ToLower(name)+".key")
}

func (s *CertStore) Write(name string, crt *x509.Certificate, key *rsa.PrivateKey) error {
	if err := ioutil.WriteFile(s.CertFile(name), cert.EncodeCertPEM(crt), 0755); err != nil {
		return fmt.Errorf("Failed to write `%s`. Reason. %v.", s.CertFile(name), err)
	}
	if err := ioutil.WriteFile(s.KeyFile(name), cert.EncodePrivateKeyPEM(key), 0600); err != nil {
		return fmt.Errorf("Failed to write `%s`. Reason. %v.", s.KeyFile(name), err)
	}
	return nil
}

func (s *CertStore) Read(name string) (*x509.Certificate, *rsa.PrivateKey, error) {
	crtBytes, err := ioutil.ReadFile(s.CertFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read certificate `%s`.Reason: %v.", s.CertFile(name), err)
	}
	crt, err := cert.ParseCertsPEM(crtBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse certificate `%s`.Reason: %v.", s.CertFile(name), err)
	}

	keyBytes, err := ioutil.ReadFile(s.KeyFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read private key `%s`.Reason: %v.", s.KeyFile(name), err)
	}
	key, err := cert.ParsePrivateKeyPEM(keyBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to parse private key `%s`.Reason: %v.", s.KeyFile(name), err)
	}
	return crt[0], key.(*rsa.PrivateKey), nil
}

func (s *CertStore) ReadBytes(name string) ([]byte, []byte, error) {
	crtBytes, err := ioutil.ReadFile(s.CertFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read certificate `%s`.Reason: %v.", s.CertFile(name), err)
	}

	keyBytes, err := ioutil.ReadFile(s.KeyFile(name))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to read private key `%s`.Reason: %v.", s.KeyFile(name), err)
	}
	return crtBytes, keyBytes, nil
}
