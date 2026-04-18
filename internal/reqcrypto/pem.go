package reqcrypto

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// LoadRSAPublicKeyFromFile loads an RSA public key from PEM (PKCS#1, PKIX, or X.509 certificate).
func LoadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM block in public key file")
	}
	switch block.Type {
	case "RSA PUBLIC KEY":
		return x509.ParsePKCS1PublicKey(block.Bytes)
	case "PUBLIC KEY":
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := pub.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA public key, got %T", pub)
		}
		return rs, nil
	case "CERTIFICATE":
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := cert.PublicKey.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate public key is %T, want RSA", cert.PublicKey)
		}
		return rs, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}

// LoadRSAPrivateKeyFromFile loads an RSA private key from PEM (PKCS#1 or PKCS#8).
func LoadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("no PEM block in private key file")
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rs, ok := k.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("expected RSA private key, got %T", k)
		}
		return rs, nil
	default:
		return nil, fmt.Errorf("unsupported PEM type %q", block.Type)
	}
}
