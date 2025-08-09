package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// NewConfigurableClient creates an http.Client that can be configured with a custom CA certificate.
// If the caCertPath is empty, it returns the default http.Client.
// This allows for secure connections in production while enabling trust for custom CAs in development.
func NewConfigurableClient(caCertPath string) *http.Client {
	if caCertPath == "" {
		// In production or when no custom CA is needed, use the default client.
		zap.L().Info("Using default system HTTP client")
		return http.DefaultClient
	}

	zap.L().Info("Attempting to load custom CA certificate", zap.String("path", caCertPath))

	// Load the custom CA certificate.
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		// Use Fatal to stop the application if the configured cert cannot be read.
		zap.L().Fatal("Failed to read custom CA certificate file. Please check the path.", zap.String("path", caCertPath), zap.Error(err))
	}

	// Get the system's certificate pool.
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Append the custom CA certificate to the system pool.
	if ok := rootCAs.AppendCertsFromPEM(caCert); !ok {
		// Use Fatal to stop the application if the configured cert is invalid.
		zap.L().Fatal("Failed to append custom CA certificate to pool. Please check the certificate format.", zap.String("path", caCertPath))
	} else {
		// Log the subject of the loaded certificate for verification.
		// This helps confirm if the correct certificate is being used.
		cert, _ := x509.ParseCertificate(caCert)
		if cert != nil {
			zap.L().Info("Successfully loaded and appended custom CA certificate", zap.String("subject", cert.Subject.String()))
		}
	}

	// Create a new TLS configuration with the combined certificate pool.
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	// Create a new HTTP transport with the custom TLS configuration.
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	zap.L().Info("Successfully created HTTP client with custom CA certificate")
	return &http.Client{Transport: transport}
}
