package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	certPath := flag.String("cert", "server.crt", "cert file")
	keyPath := flag.String("key", "server.key", "key file")
	hosts := flag.String("host", "localhost", "comma separated hosts (DNS or IP)")
	flag.Parse()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	serial, _ := rand.Int(rand.Reader, big.NewInt(1<<62))
	hostList := strings.Split(*hosts, ",")
	var dnsNames []string
	var ipAddrs []net.IP
	for _, h := range hostList {
		h = strings.TrimSpace(h)
		if ip := net.ParseIP(h); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		} else if h != "" {
			dnsNames = append(dnsNames, h)
		}
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "Go App Manager"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		DNSNames:     dnsNames,
		IPAddresses:  ipAddrs,
		IsCA:         true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}

	certOut, err := os.Create(*certPath)
	if err != nil {
		panic(err)
	}
	defer certOut.Close()
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyOut, err := os.Create(*keyPath)
	if err != nil {
		panic(err)
	}
	defer keyOut.Close()
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
}
