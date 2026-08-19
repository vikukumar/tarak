// Package security provides TLS certificate management for the Tarak cluster.
//
// Phase 1 implements a complete self-signed PKI:
//   - A root Certificate Authority (CA) is generated on first init and stored on disk.
//   - Server certificates are signed by the CA for the API server.
//   - Client certificates are signed by the CA for:
//       - The cluster admin (used by tarakctl)
//       - Internal components (scheduler, controller, agent)
//   - All certificates embed the standard Kubernetes-compatible CN/O fields so that
//     the authentication middleware can extract user/group information.
//
// Certificate types and their CN/O conventions:
//
//	CA:             CN=tarak-ca
//	API Server:     CN=tarak-apiserver, SAN=localhost,127.0.0.1,<nodeIP>
//	Admin:          CN=tarak-admin, O=system:masters
//	Scheduler:      CN=system:kube-scheduler
//	Controller:     CN=system:kube-controller-manager
//	Agent:          CN=system:node:<nodeName>, O=system:nodes
package security

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertificateAuthority holds the root CA certificate and key.
type CertificateAuthority struct {
	// Certificate is the DER-encoded root CA certificate.
	Certificate *x509.Certificate
	// CertPEM is the PEM-encoded root CA certificate (for writing to disk / TLS config).
	CertPEM []byte
	// PrivateKey is the CA private key (P-256 ECDSA).
	PrivateKey *ecdsa.PrivateKey
	// KeyPEM is the PEM-encoded CA private key.
	KeyPEM []byte
}

// CertKeyPair is a signed certificate and its matching private key.
type CertKeyPair struct {
	// Certificate is the DER-encoded leaf certificate.
	Certificate *x509.Certificate
	// CertPEM is the PEM-encoded certificate (+ CA chain).
	CertPEM []byte
	// PrivateKey is the leaf private key.
	PrivateKey *ecdsa.PrivateKey
	// KeyPEM is the PEM-encoded private key.
	KeyPEM []byte
}

// PKIDirectory is the layout of the on-disk PKI directory.
const (
	fileCAcert     = "ca.crt"
	fileCAkey      = "ca.key"
	fileServerCert = "apiserver.crt"
	fileServerKey  = "apiserver.key"
	fileAdminCert  = "admin.crt"
	fileAdminKey   = "admin.key"
)

// ─── CA generation ────────────────────────────────────────────────────────────

// GenerateCA creates a new root Certificate Authority.
// The CA certificate is valid for 10 years.
func GenerateCA() (*CertificateAuthority, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate CA key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return nil, fmt.Errorf("generate CA serial: %w", err)
	}

	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   "tarak-ca",
			Organization: []string{"Tarak"},
		},
		NotBefore:             now.Add(-10 * time.Second), // small back-date to handle clock skew
		NotAfter:              now.Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLenZero:        false,
		MaxPathLen:            2,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, fmt.Errorf("create CA cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse CA cert: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM, err := encodeECKey(key)
	if err != nil {
		return nil, err
	}

	return &CertificateAuthority{
		Certificate: cert,
		CertPEM:     certPEM,
		PrivateKey:  key,
		KeyPEM:      keyPEM,
	}, nil
}

// LoadCA loads a CA from PEM-encoded certificate and key bytes.
func LoadCA(certPEM, keyPEM []byte) (*CertificateAuthority, error) {
	cert, key, err := decodePEMPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("load CA: %w", err)
	}
	return &CertificateAuthority{
		Certificate: cert,
		CertPEM:     certPEM,
		PrivateKey:  key,
		KeyPEM:      keyPEM,
	}, nil
}

// ─── Certificate signing ──────────────────────────────────────────────────────

// ServerCertOptions configures generation of an API server certificate.
type ServerCertOptions struct {
	// CommonName is the CN of the certificate.
	CommonName string
	// SANs is the list of Subject Alternative Names (hostnames and IPs).
	SANs []string
	// ValidFor is the certificate lifetime (default 1 year).
	ValidFor time.Duration
}

// SignServerCert creates a server TLS certificate signed by the CA.
func (ca *CertificateAuthority) SignServerCert(opts ServerCertOptions) (*CertKeyPair, error) {
	if opts.ValidFor == 0 {
		opts.ValidFor = 365 * 24 * time.Hour
	}
	cn := opts.CommonName
	if cn == "" {
		cn = "tarak-apiserver"
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate server key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return nil, err
	}

	var dnsNames []string
	var ipAddrs []net.IP
	for _, san := range opts.SANs {
		if ip := net.ParseIP(san); ip != nil {
			ipAddrs = append(ipAddrs, ip)
		} else {
			dnsNames = append(dnsNames, san)
		}
	}
	// Always include localhost / 127.0.0.1.
	if len(dnsNames) == 0 {
		dnsNames = []string{"localhost"}
	}
	if !containsIP(ipAddrs, net.ParseIP("127.0.0.1")) {
		ipAddrs = append(ipAddrs, net.ParseIP("127.0.0.1"))
	}
	if !containsIP(ipAddrs, net.ParseIP("::1")) {
		ipAddrs = append(ipAddrs, net.ParseIP("::1"))
	}

	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   cn,
			Organization: []string{"Tarak"},
		},
		NotBefore:   now.Add(-10 * time.Second),
		NotAfter:    now.Add(opts.ValidFor),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
		IPAddresses: ipAddrs,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Certificate, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create server cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse server cert: %w", err)
	}

	// Include CA cert in PEM bundle for clients that need the full chain.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certPEM = append(certPEM, ca.CertPEM...)
	keyPEM, err := encodeECKey(key)
	if err != nil {
		return nil, err
	}

	return &CertKeyPair{
		Certificate: cert,
		CertPEM:     certPEM,
		PrivateKey:  key,
		KeyPEM:      keyPEM,
	}, nil
}

// ClientCertOptions configures generation of a client certificate.
type ClientCertOptions struct {
	// CommonName is the CN — used as the username by the auth middleware.
	CommonName string
	// Organizations are the O fields — used as group names by the auth middleware.
	Organizations []string
	// ValidFor is the certificate lifetime (default 1 year).
	ValidFor time.Duration
}

// SignClientCert creates a client TLS certificate signed by the CA.
func (ca *CertificateAuthority) SignClientCert(opts ClientCertOptions) (*CertKeyPair, error) {
	if opts.ValidFor == 0 {
		opts.ValidFor = 365 * 24 * time.Hour
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate client key: %w", err)
	}

	serial, err := randomSerial()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:   opts.CommonName,
			Organization: opts.Organizations,
		},
		NotBefore:   now.Add(-10 * time.Second),
		NotAfter:    now.Add(opts.ValidFor),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, ca.Certificate, &key.PublicKey, ca.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create client cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, fmt.Errorf("parse client cert: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM, err := encodeECKey(key)
	if err != nil {
		return nil, err
	}

	return &CertKeyPair{
		Certificate: cert,
		CertPEM:     certPEM,
		PrivateKey:  key,
		KeyPEM:      keyPEM,
	}, nil
}

// ─── PKI directory I/O ───────────────────────────────────────────────────────

// WritePKI writes a complete PKI to a directory:
//   - ca.crt, ca.key
//   - apiserver.crt, apiserver.key
//   - admin.crt, admin.key
func WritePKI(dir string, ca *CertificateAuthority, server, admin *CertKeyPair) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create PKI dir %q: %w", dir, err)
	}

	pairs := []struct{ name string; data []byte }{
		{fileCAcert, ca.CertPEM},
		{fileCAkey, ca.KeyPEM},
		{fileServerCert, server.CertPEM},
		{fileServerKey, server.KeyPEM},
		{fileAdminCert, admin.CertPEM},
		{fileAdminKey, admin.KeyPEM},
	}
	for _, p := range pairs {
		if err := os.WriteFile(filepath.Join(dir, p.name), p.data, 0600); err != nil {
			return fmt.Errorf("write %q: %w", p.name, err)
		}
	}
	return nil
}

// LoadPKI loads the CA, server cert, and admin cert from a PKI directory.
func LoadPKI(dir string) (ca *CertificateAuthority, server, admin *CertKeyPair, err error) {
	readFile := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, name))
	}

	caCertPEM, err := readFile(fileCAcert)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read CA cert: %w", err)
	}
	caKeyPEM, err := readFile(fileCAkey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read CA key: %w", err)
	}
	ca, err = LoadCA(caCertPEM, caKeyPEM)
	if err != nil {
		return nil, nil, nil, err
	}

	serverCertPEM, err := readFile(fileServerCert)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read server cert: %w", err)
	}
	serverKeyPEM, err := readFile(fileServerKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read server key: %w", err)
	}
	serverCert, serverKey, err := decodePEMPair(serverCertPEM, serverKeyPEM)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode server pair: %w", err)
	}
	server = &CertKeyPair{Certificate: serverCert, CertPEM: serverCertPEM, PrivateKey: serverKey, KeyPEM: serverKeyPEM}

	adminCertPEM, err := readFile(fileAdminCert)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read admin cert: %w", err)
	}
	adminKeyPEM, err := readFile(fileAdminKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read admin key: %w", err)
	}
	adminCert, adminKey, err := decodePEMPair(adminCertPEM, adminKeyPEM)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("decode admin pair: %w", err)
	}
	admin = &CertKeyPair{Certificate: adminCert, CertPEM: adminCertPEM, PrivateKey: adminKey, KeyPEM: adminKeyPEM}

	return ca, server, admin, nil
}

// ─── x509 pool helpers ────────────────────────────────────────────────────────

// CertPool returns an *x509.CertPool containing only the CA certificate.
func (ca *CertificateAuthority) CertPool() (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca.CertPEM) {
		return nil, fmt.Errorf("CA CertPool: failed to append CA cert")
	}
	return pool, nil
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// randomSerial generates a cryptographically random 128-bit serial number.
func randomSerial() (*big.Int, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("random serial: %w", err)
	}
	return new(big.Int).SetBytes(b), nil
}

// encodeECKey PEM-encodes an ECDSA private key.
func encodeECKey(key *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("marshal EC key: %w", err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: der}), nil
}

// decodePEMPair decodes a PEM certificate + PEM EC private key pair.
func decodePEMPair(certPEM, keyPEM []byte) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, fmt.Errorf("decodePEMPair: no CERTIFICATE block found")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse certificate: %w", err)
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "EC PRIVATE KEY" {
		return nil, nil, fmt.Errorf("decodePEMPair: no EC PRIVATE KEY block found")
	}
	key, err := x509.ParseECPrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse EC private key: %w", err)
	}

	return cert, key, nil
}

// containsIP reports whether the slice contains ip.
func containsIP(ips []net.IP, ip net.IP) bool {
	for _, x := range ips {
		if x.Equal(ip) {
			return true
		}
	}
	return false
}
