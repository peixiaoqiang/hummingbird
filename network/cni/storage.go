package main

import (
	"context"
	"errors"
	"github.com/coreos/etcd/client"
		"log"
	"os"
	"strings"
	"time"
	)

type EtcdHelper struct {
	HeaderTimeoutPerRequest time.Duration
	Client                  client.Client
}


func (IpamS *IpamConfig) etcdConn() (EtcdConn EtcdHelper) {

	//tlsInfo := transport.TLSInfo{
	//	CertFile: IpamS.Ipam.CertFile,
	//	KeyFile:  IpamS.Ipam.KeyFile,
	//	TrustedCAFile:   IpamS.Ipam.CAFile,
	//}
	//
	//t, err := transport.NewTransport(tlsInfo, time.Second)

	var etcdServerList []string = strings.Split(IpamS.Ipam.Etcdcluster, ",")
	cli, err := client.New(client.Config{
		Endpoints:               etcdServerList,
		HeaderTimeoutPerRequest: 1 * time.Second,
	//	Transport: t,
	//	Username:  IpamS.Ipam.Username,
	//	Password:  IpamS.Ipam.Password,
	})

	if err != nil {
		log.Println("connect failed, err:", err)
		os.Exit(-1)

	}

	return EtcdHelper{
		HeaderTimeoutPerRequest: 1 * time.Second,
		Client:                  cli,
	}

}

func IsKeyExist(Rang *map[string]string, Key string) error {
	if _, ok := (*Rang)[Key+"RangeStart"]; !ok {
		return errors.New("ETCD Lack rangeStart")
	}

	if _, ok := (*Rang)[Key+"RangeEnd"]; !ok {
		return errors.New("ETCD Lack rangeEnd")

	}
	return nil
}

func (Cli EtcdHelper) setKey(keyRoad string, key string, containerID string) error {
	kapi := client.NewKeysAPI(Cli.Client)
	_, err := kapi.Set(context.Background(), keyRoad+key, containerID,&client.SetOptions{PrevExist:client.PrevNoExist})
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (Cli EtcdHelper) getKey(key string) (NodesInfo *map[string]string) {
	kapi := client.NewKeysAPI(Cli.Client)
	resp, err := kapi.Get(context.Background(), key, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Fatal(err)
		return
	}
	skydnsNodesInfo := make(map[string]string)
	getAllNode(resp.Node, skydnsNodesInfo)
	return &skydnsNodesInfo
}

func (Cli EtcdHelper) delKey(key string) error {
	kapi := client.NewKeysAPI(Cli.Client)

	//_, err := kapi.Delete(context.Background(), "/foo", &client.DeleteOptions{PrevValue: "bar"})
	_, err := kapi.Delete(context.Background(), key, &client.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil

}

func getAllNode(rootNode *client.Node, nodesInfo map[string]string) {
	if !rootNode.Dir {
		nodesInfo[rootNode.Key] = rootNode.Value
		return
	}
	for node := range rootNode.Nodes {
		getAllNode(rootNode.Nodes[node], nodesInfo)
	}
}
