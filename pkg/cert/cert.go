package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const (
	ServerCertFile = "/tmp/mtf/cert/server.crt"
	ServerKeyFile  = "/tmp/mtf/cert/server.key"
)

func WriteCert(ck *CertKey) error {
	dir := filepath.Dir(ServerCertFile)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	if err := ioutil.WriteFile(ServerCertFile, ck.Cert, 0665); err != nil {
		return err
	}
	if err := ioutil.WriteFile(ServerKeyFile, ck.Key, 0665); err != nil {
		return err
	}
	return nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

type CertKey struct {
	Cert []byte
	Key  []byte
}

func GenCert(hosts []string) (*CertKey, error) {
	var priv interface{}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number: %s", err)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme"},
		},
		NotBefore: now,
		NotAfter:  now.Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts = append(hosts, []string{"localhost", "host.docker.internal"}...)
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	template.IsCA = true
	template.KeyUsage |= x509.KeyUsageCertSign

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return nil, fmt.Errorf("Failed to create certificate: %s", err)
	}

	var certBuff bytes.Buffer
	if err := pem.Encode(&certBuff, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return nil, fmt.Errorf("failed to write data to cert.pem: %s", err)
	}

	var keyBuff bytes.Buffer
	if err := pem.Encode(&keyBuff, pemBlockForKey(priv)); err != nil {
		return nil, fmt.Errorf("failed to write data to key.pem: %s", err)
	}

	return &CertKey{
		Cert: certBuff.Bytes(),
		Key:  keyBuff.Bytes(),
	}, nil
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
