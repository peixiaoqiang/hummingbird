package main

import (
	"crypto/tls"
	"flag"
	"net/http"

	"github.com/golang/glog"

	"github.com/TalkingData/hummingbird/pkg/spark"
)

type Config struct {
	CertFile string
	KeyFile  string
}

func (c *Config) addFlags() {
	flag.StringVar(&c.CertFile, "tls-cert-file", c.CertFile, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert).")
	flag.StringVar(&c.KeyFile, "tls-private-key-file", c.KeyFile, ""+
		"File containing the default x509 private key matching --tls-cert-file.")
}

func configTLS(config Config) *tls.Config {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		glog.Fatal(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
}

func main() {
	var config Config
	config.addFlags()
	flag.Parse()
	defer glog.Flush()

	http.HandleFunc("/mutating-pods", spark.ServeMutatePods)
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: configTLS(config),
	}
	glog.Info("start server on :443")
	err := server.ListenAndServeTLS("", "")
	if err != nil {
		glog.Errorf("fail to start server: %v", err)

	}
}
