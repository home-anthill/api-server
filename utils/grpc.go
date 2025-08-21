package utils

import (
	"api-server/customerrors"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func BuildSecurityDialOption() (grpc.DialOption, bool, error) {
	var securityDialOption grpc.DialOption
	if os.Getenv("GRPC_TLS") == "true" {
		tlsCredentials, errTLS := LoadTLSCredentials()
		if errTLS != nil {
			return nil, false, customerrors.Wrap(http.StatusInternalServerError, errTLS, "loadTLSCredentials cannot read certificates")
		}
		securityDialOption = grpc.WithTransportCredentials(tlsCredentials)
		return securityDialOption, true, nil
	}

	// if security is not enabled, use the insecure version
	securityDialOption = grpc.WithTransportCredentials(insecure.NewCredentials())
	return securityDialOption, false, nil
}

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(os.Getenv("CERT_FOLDER_PATH") + "/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
