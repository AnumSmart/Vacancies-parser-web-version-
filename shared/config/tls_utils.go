// описание работы с TLS сертификатами (только для индививуального использования HTTPS у сервера. т.е. API Gareway типа nginx не используется)
package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// LoadTLSCertificate загружает и проверяет TLS сертификат
func LoadTLSCertificate(certFile, keyFile string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Проверяем, что приватный ключ соответствует сертификату
	if _, err := x509.ParseCertificate(cert.Certificate[0]); err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// CheckCertificateValidity проверяет срок действия сертификата
func CheckCertificateValidity(certFile string) error {
	data, err := os.ReadFile(certFile)
	if err != nil {
		return fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return fmt.Errorf("failed to decode PEM block from certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	now := time.Now()

	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from: %s)", cert.NotBefore.Format(time.RFC3339))
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired at: %s)", cert.NotAfter.Format(time.RFC3339))
	}

	// Предупреждение, если срок действия истекает в ближайшие 30 дней
	if time.Until(cert.NotAfter) < 30*24*time.Hour {
		daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
		return fmt.Errorf("certificate expires soon (in %d days)", daysLeft)
	}

	return nil
}

/*
// CreateTLSConfig создает конфигурацию TLS для HTTP сервера
func (c *ServerConfig) CreateTLSConfig() (*tls.Config, error) {
	if !c.EnableTLS {
		return nil, nil
	}

	// Загружаем сертификат
	cert, err := LoadTLSCertificate(c.TLSCertFile, c.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Проверяем валидность сертификата
	if err := CheckCertificateValidity(c.TLSCertFile); err != nil {
		// Здесь можно логировать предупреждение, но не прерывать запуск
		// для development окружения
		fmt.Printf("Certificate warning: %v\n", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12, // Минимум TLS 1.2 для безопасности
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}, nil
}
*/
