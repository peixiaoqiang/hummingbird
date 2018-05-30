package main

import (
	"encoding/json"
	"github.com/containernetworking/cni/pkg/skel"
	"github.com/containernetworking/cni/pkg/types"
	"github.com/containernetworking/cni/pkg/types/current"
	"github.com/containernetworking/cni/pkg/version"
	"log"
	"os"
	"net"
	"strings"
	"flag"
		"io/ioutil"
	)

type EtcdConfig struct {
	Etcdcluster      string `json:etcd server地址`
	Nodenetwork      string `json:node network`
	Alreadyusedip    string `json:container AvailableIp`
	Containernetwork string `json:containernetr`
	Routes           string `json:routes`
	Dns              string `json:dns`
}

type IpamConfig struct {
	Ipam       EtcdConfig
	Name       string `json:"name"`
	CNIVersion string `json:"cniVersion"`
}

func (IpamS *IpamConfig) Load(bytes []byte) error {
	err := json.Unmarshal(bytes, IpamS)
	if err != nil {
		log.Println("error in translating,", err.Error())

	}

	return nil
	//fmt.Println("type:", reflect.TypeOf(IpamS.Ipam.Etcdcluster))
}


func main(){

	Config := IpamConfig{}
	//用来初始化etcd里中的配置项，如果默认不用来初始化IP，将执行正常分配IP的流程
	init := flag.String("init","","init")
	RangeStart := flag.String("start","","172.20.0.140")
	RangeEnd := flag.String("end","","172.20.0.150")
	SubNet := flag.String("subnet","","172.20.0.140/17")
	GateWay := flag.String("gateway","","172.20.127.254")
	ConfiFile := flag.String("config","","/etc/cni/net.d/10-macvlan.conf")
	flag.Parse()
	if *init == "init"{
		if contents, err:= ioutil.ReadFile(*ConfiFile);err == nil{
			Config.Load(contents)
		}else{
			log.Fatalf("Lack of configuration files")
		}

		Cli := Config.etcdConn()
		start := net.ParseIP(*RangeStart)
		if start == nil{
			log.Fatalf("Incorrect rangStart address")
		} else{
			err := Cli.setKey(Config.Ipam.Containernetwork,"RangeStart",*RangeStart)
			if err != nil{
				log.Fatalf("Create key RangeStart failure")
			}
		}

		end := net.ParseIP(*RangeEnd)
		if end == nil{
			log.Fatalf("Incorrect rangEnd address")
		}else{
			err := Cli.setKey(Config.Ipam.Containernetwork,"RangeEnd",*RangeEnd)
			if err != nil{
				log.Fatalf("Create key rangEnd failure")
			}
		}

		_, _, err := net.ParseCIDR(*SubNet)
		if err !=nil {
			log.Fatal("Incorrect SubNet SubNet")
		}else{
			err := Cli.setKey(Config.Ipam.Containernetwork,"SubNet",*SubNet)
			if err != nil{
				log.Fatalf("Create key SubNet failure")
			}
		}

		gateway := net.ParseIP(*GateWay)
		if gateway == nil{
			log.Fatalf("Incorrect gateway address")
		}else{
			err := Cli.setKey(Config.Ipam.Containernetwork,"GateWay",*GateWay)
			if err != nil{
				log.Fatalf("Create key GateWay failure")
			}
		}

	}
	skel.PluginMain(cmdAdd, cmdDel, version.All)
}

func cmdAdd(args *skel.CmdArgs) error {
	//f, err := os.OpenFile("/tmp/ipam.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatalf("error opening file: %v", err)
	//}
	//defer f.Close()
	//
	//log.SetOutput(f)
	//log.Println("start ipam...")


	//读取配置
	Config := IpamConfig{}
	Config.Load(args.StdinData)
	//连接etcd
	Cli := Config.etcdConn()

	//NodeRang := Cli.getKey(Config.Ipam.Nodenetwork)
	//err := IsKeyExist(NodeRang,Config.Ipam.Nodenetwork)
	//
	//if err != nil{
	//	log.Println(err)
	//	os.Exit(-1)
	//}
	ContainerRange := Cli.getKey(Config.Ipam.Containernetwork)
	err := IsKeyExist(ContainerRange, Config.Ipam.Containernetwork)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	//获取pod name
	var podName string
	Args := strings.Split(args.Args,";")
	for _, as := range Args{
		if strings.Contains(as,"K8S_POD_NAME"){
			pods := strings.Split(as,"=")
			podName = pods[len(pods)-1]

		}
	}

	//从etcd库中pod_name是否已经存在Ip，如果存在，则不需要再从新分配IP,并且将新的容器ID写入etcd
	var AvailableIp net.IP
	if len(podName) > 0 {
		existPodName := Cli.getKey(Config.Ipam.Alreadyusedip + "podname/")
		log.Println(podName)
		if existIp, ok := (*existPodName)[Config.Ipam.Alreadyusedip + "podname/"+podName]; ok {
			AvailableIp = net.ParseIP(existIp)
			err = Cli.setKey(Config.Ipam.Alreadyusedip, AvailableIp.String(), args.ContainerID)
			if err != nil {
				log.Println(err)
			}
		}
	}

	//从etcd获取容器IP范围
	var ContainerR *Range
	ContainerR = &Range{}
	ContainerR.Container(ContainerRange, Config.Ipam.Containernetwork)
	//效验IP地址范围是否正确
	ContainerR.Canonicalize()
	//从etcd获取目前已经使用掉的IP
	AlreadUsedIp := Cli.getKey(Config.Ipam.Alreadyusedip)
	//获取目前可用的IP地址
	IpList, err := Hosts((*ContainerRange)[Config.Ipam.Containernetwork+"SubNet"])
	if err != nil {
		log.Println("IP地址范围错误")
	}

	//如果etcd中不存在IP则分配IP
	if AvailableIp == nil {
		AvailableIpList := ContainerR.RangeSet(AlreadUsedIp, &IpList, Config.Ipam.Alreadyusedip)
		//将获取到的IP提交到ETCD库中

		for _, Ip := range AvailableIpList {
			if len(Ip.String()) > 0 {
				err = Cli.setKey(Config.Ipam.Alreadyusedip, Ip.String(), args.ContainerID)
			}
			if err == nil {
				AvailableIp = Ip
				err = Cli.setKey(Config.Ipam.Alreadyusedip+"podname/", podName, Ip.String())
				if err != nil {
					log.Println(err)
				}
				break
			}
		}
		if len(AvailableIp.String()) <= 0 {
			log.Println("没有可用Ip")
		}
	}
	//返回cni相关
	result := &current.Result{}
	//返回cni 版本
	result.CNIVersion = Config.CNIVersion
	//返给cni ip
	IPs := &current.IPConfig{}
	IPs.Gateway = ContainerR.Gateway
	IPs.Version = "4"
	IPs.Address.IP = AvailableIp

	IPs.Address.Mask = ContainerR.Subnet.Mask
	result.IPs = append(result.IPs, IPs)
	//获取dns配置
	dnsEtcdConfig := Cli.getKey(Config.Ipam.Dns)
	result.DNS = GetDns(dnsEtcdConfig, &Config)

	//自定义容器路由规则
	//routeEtcdConfig := Cli.getKey(Config.Ipam.Routes)
	//GetRoute(routeEtcdConfig, &Config)
	//_, dstmask, err := net.ParseCIDR("0.0.0.0/0")
	//Routes := &types.Route{}
	//Routes.GW = nil
	//Routes.Dst.IP = dstmask.IP
	//Routes.Dst.Mask = dstmask.Mask
	//result.Routes = append(result.Routes,Routes)
	return types.PrintResult(result, Config.CNIVersion)

}

func cmdDel(args *skel.CmdArgs) error {
	//etcd KEY路劲
	//读取配置
	//f, err := os.OpenFile("/tmp/stopipam.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatalf("error opening file: %v", err)
	//}
	//defer f.Close()
	//
	//log.SetOutput(f)

	Config := IpamConfig{}
	Config.Load(args.StdinData)
	Cli := Config.etcdConn()
	AlreadUsedIp := Cli.getKey(Config.Ipam.Alreadyusedip)

	//连接etcd
	for k,v := range *AlreadUsedIp {
		if v == args.ContainerID{
			log.Println(args.ContainerID)
			log.Println("delete key")
			Cli.delKey(k)
		}
	}

	//根据容器ID查找IP
	Key := ContainerSearch(AlreadUsedIp, args.ContainerID)
	if len(Key) > 0 {
		log.Println("delete ip key")
		Cli.delKey(Key)
	}
	//获取pod name
	var podName string
	Args := strings.Split(args.Args,";")
	for _, as := range Args{
		if strings.Contains(as,"K8S_POD_NAME"){
			pods := strings.Split(as,"=")
			podName = pods[len(pods)-1]

		}
	}
	//删除podname
	if len(podName) > 0 {
		log.Println("delete" + Config.Ipam.Alreadyusedip + "podname/" + podName)
		Cli.delKey(Config.Ipam.Alreadyusedip + "podname/" + podName)
	}

	return nil
}
