package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/kelvinzer0/imail/internal/tools"
)

func makeRsa() ([]byte, []byte, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	publickey := &privatekey.PublicKey
	Priv := x509.MarshalPKCS1PrivateKey(privatekey)
	Pub, err := x509.MarshalPKIXPublicKey(publickey)
	if err != nil {
		return nil, nil, err
	}
	return Priv, Pub, nil
}

func CheckDomainARecord(domain string) error {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return err
	}

	ip, err := tools.GetPublicIP()
	if err != nil {
		return err
	}

	var isFind = false
	for _, fIp := range findIp {
		if fIp.To4() != nil && strings.EqualFold(fIp.String(), ip) {
			isFind = true
			break
		}
	}

	if !isFind {
		return errors.New("IPv4 not configured by domain name!")
	}

	return nil
}

func CheckDomainAAAARecord(domain string) error {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return err
	}

	var isFind = false
	for _, fIp := range findIp {
		if fIp.To4() == nil && fIp.To16() != nil { // Check if it's an IPv6 address
			isFind = true
			break
		}
	}

	if !isFind {
		return errors.New("IPv6 not configured by domain name!")
	}

	return nil
}

func MakeDkimFile(path, domain string) (string, error) {
	priFile := fmt.Sprintf("%s/dkim/%s/default.private", path, domain)
	defalutTextFile := fmt.Sprintf("%s/dkim/%s/default.txt", path, domain)
	defalutValFile := fmt.Sprintf("%s/dkim/%s/default.val", path, domain)

	if tools.IsExist(priFile) {
		pubContent, err := tools.ReadFile(defalutTextFile)
		if err != nil {
			return "", err
		}
		return pubContent, nil
	}

	Priv, Pub, err := makeRsa()
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: Priv,
	}

	// pri := b64.StdEncoding.EncodeToString(Priv)
	file, err := os.Create(priFile)
	if err != nil {
		return "", err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return "", err
	}

	pub := b64.StdEncoding.EncodeToString(Pub)
	pubContent := fmt.Sprintf("default._domainkey\tIN\tTXT\t(\r\nv=DKIM1;k=rsa;p=%s\r\n)\r\n----- DKIM key default for %s", pub, domain)

	err = tools.WriteFile(defalutTextFile, pubContent)
	if err != nil {
		return "", err
	}
	err = tools.WriteFile(defalutValFile, fmt.Sprintf("v=DKIM1;k=rsa;p=%s", pub))
	if err != nil {
		return "", err
	}

	return pubContent, nil
}

func MakeDkimConfFile(path, domain string) (string, error) {
	pDir := fmt.Sprintf("%s/dkim", path)
	if b := tools.IsExist(pDir); !b {
		err := os.MkdirAll(pDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	pathDir := fmt.Sprintf("%s/dkim/%s", path, domain)
	if b := tools.IsExist(pathDir); !b {
		err := os.MkdirAll(pathDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return MakeDkimFile(path, domain)
}

func GetDomainDkimVal(path, domain string) (string, error) {
	_, err := MakeDkimConfFile(path, domain)
	if err != nil {
		return "", err
	}
	defalutValFile := fmt.Sprintf("%s/dkim/%s/default.val", path, domain)
	pubContentRecord, err := tools.ReadFile(defalutValFile)
	return pubContentRecord, err
}
