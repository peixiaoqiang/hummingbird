/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package factory

import (
	"net"
	"net/http"
	"time"

	etcd2client "github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"

	"github.com/TalkingData/hummingbird/pkg/storage"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend"
	"github.com/TalkingData/hummingbird/pkg/storage/storagebackend/etcd"
	utilnet "github.com/TalkingData/hummingbird/pkg/utils/net"
)

func newETCD2Storage(c storagebackend.Config) (storage.Interface, DestroyFunc, error) {
	tr, err := newTransportForETCD2(c.CertFile, c.KeyFile, c.CAFile)
	if err != nil {
		return nil, nil, err
	}
	client, err := newETCD2Client(tr, c.ServerList)
	if err != nil {
		return nil, nil, err
	}
	s := etcd.NewEtcdStorage(client, c.Prefix, c.Quorum)

	return s, tr.CloseIdleConnections, nil
}

func newETCD2Client(tr *http.Transport, serverList []string) (etcd2client.Client, error) {
	cli, err := etcd2client.New(etcd2client.Config{
		Endpoints: serverList,
		Transport: tr,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func newTransportForETCD2(certFile, keyFile, caFile string) (*http.Transport, error) {
	info := transport.TLSInfo{
		CertFile:      certFile,
		KeyFile:       keyFile,
		TrustedCAFile: caFile,
	}
	cfg, err := info.ClientConfig()
	if err != nil {
		return nil, err
	}
	// Copied from etcd.DefaultTransport declaration.
	// TODO: Determine if transport needs optimization
	tr := utilnet.SetTransportDefaults(&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: 500,
		TLSClientConfig:     cfg,
	})
	return tr, nil
}
