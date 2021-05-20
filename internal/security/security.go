//
// Copyright Strimzi authors.
// License: Apache License 2.0 (see the file LICENSE or http://apache.org/licenses/LICENSE-2.0.html).
//

// Package security defining some security related tools
package security

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
	"github.com/strimzi/strimzi-canary/internal/config"
)

func NewTLSConfig(canaryConfig *config.CanaryConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	var e error
	// with TLS enabled, the System certs pool is used by default but adding the CA certificate if provided
	if tlsConfig.RootCAs, e = x509.SystemCertPool(); e != nil {
		glog.Infof("Error on setting the System certs pool as root CAs: %v", e)
	}

	if canaryConfig.TLSCACert != "" {
		if caCert, err := loadCertKey(config.TLSCACertEnvVar, canaryConfig.TLSCACert); err == nil {
			if tlsConfig.RootCAs == nil {
				tlsConfig.RootCAs = x509.NewCertPool()
			}
			if !tlsConfig.RootCAs.AppendCertsFromPEM(caCert) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if canaryConfig.TLSClientCert != "" && canaryConfig.TLSClientKey != "" {
		var clientCert, clientKey []byte
		var err error
		var cert tls.Certificate

		if clientCert, err = loadCertKey(config.TLSClientCertEnvVar, canaryConfig.TLSClientCert); err != nil {
			return nil, err
		}
		if clientKey, err = loadCertKey(config.TLSClientKeyEnvVar, canaryConfig.TLSClientKey); err != nil {
			return nil, err
		}
		if cert, err = tls.X509KeyPair(clientCert, clientKey); err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	tlsConfig.InsecureSkipVerify = canaryConfig.TLSInsecureSkipVerify

	return tlsConfig, nil
}

func loadCertKey(config string, value string) ([]byte, error) {
	var bytes []byte
	// first check if the config is providing a file path to the certificate/key
	if _, err := os.Stat(value); err == nil {
		if bytes, err = ioutil.ReadFile(value); err != nil {
			return nil, err
		}
		glog.Infof("%s loaded from file path: %s", config, value)
	} else {
		// otherwise the config contains the certificate/key directly
		bytes = []byte(value)
		glog.Infof("%s loaded from configuration directly", config)
	}
	return bytes, nil
}
