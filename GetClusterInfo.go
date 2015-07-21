package main

import (
	//"time"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
)

type Net struct {
	Name string
	Rx   float64
	Tx   float64
}
type Config struct {
	Timestamp    int64
	Cpuusage     uint64
	Memmoryusage uint64
	Diskusage    uint64
	NetworkInfo  []*Net
}
type SpecialData struct {
	CpuMaxTimeStamp    int64
	CpuMax             uint64
	CpuAvg             uint64
	MemoryMaxTimeStamp int64
	MemoryMax          uint64
	MemoryAvg          uint64
	DiskMaxTimeStamp   int64
	DiskMax            uint64
	DiskAvg            uint64
	RxMaxTimeStamp     int64
	RxMax              float64
	RxAvg              float64
	TxMaxTimeStamp     int64
	TxMax              float64
	TxAvg              float64
}
type Nodestatus struct {
	//Ip             string
	DockerVersion  string
	KernelVersion  string
	OSVersion      string
	Cpucores       int
	Cpufrequency   uint64
	Memorycapacity int64
	Diskcapacity   uint64
	Spec           SpecialData
	Status         []*Config
	Index          int
}
type ClusterNetwork struct {
	TimeStamp int64
	Rx        float64
	Tx        float64
}
type FinalCluster struct {
	Status           string
	MasterRxMax      float64
	MasterRxAvg      float64
	MasterRxMaxStamp int64
	MasterTxMax      float64
	MasterTxAvg      float64
	MasterTxMaxStamp int64
	Network          []*ClusterNetwork
	Cluster          map[string]*Nodestatus
}

var (
	Ips       []string
	Cluster   map[string]*Nodestatus
	PointNums uint64
	FCluster  FinalCluster
)

/*
func NewNodestatus(ip string) *Nodestatus {
	conf := []*Config{}
	return &Nodestatus{
		Ip:        ip,
		CpuMax:    0,
		MemoryMax: 0,
		DiskMax:   0,
		RxMax:     0,
		TxMax:     0,
		Status:    conf,
		Point:     0,
	}
}
*/
func (c *Config) GetMachineInfo(ip string) error {
	machineinfo := MachineInfo{}
	client := &http.Client{}
	resp, err := client.Get("http://" + ip + ":4194/api/v1.3/machine/")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		defer resp.Body.Close()
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		_ = json.Unmarshal(buff.Bytes(), &machineinfo)
		Cluster[ip].Cpucores = machineinfo.NumCores
		Cluster[ip].Cpufrequency = machineinfo.CpuFrequency * 1000
		Cluster[ip].Memorycapacity = machineinfo.MemoryCapacity
		var Filesystem uint64 = 0
		for _, fs := range machineinfo.Filesystems {
			Filesystem += fs.Capacity
		}
		Cluster[ip].Diskcapacity = Filesystem
	}
	return err
}
func (c *Config) GetContainerInfo(ip string) error {
	containerinfo := ContainerInfo{}
	client := &http.Client{}
	resp, err := client.Get("http://" + ip + ":4194/api/v1.3/containers/")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		_ = json.Unmarshal(buff.Bytes(), &containerinfo)
		//c.Timestamp = containerinfo.Spec.CreationTime.Unix()
		c.Timestamp = containerinfo.Stats[len(containerinfo.Stats)-1].Timestamp.Unix()
		interval := containerinfo.Stats[len(containerinfo.Stats)-1].Timestamp.Sub(containerinfo.Stats[0].Timestamp)
		c.Cpuusage = uint64(float64(containerinfo.Stats[len(containerinfo.Stats)-1].Cpu.Usage.Total-containerinfo.Stats[0].Cpu.Usage.Total) / float64(interval) * 1000000000)
		var Filesystem uint64 = 0
		for _, fs := range containerinfo.Stats[0].Filesystem {
			Filesystem += fs.Usage
		}
		c.Diskusage = Filesystem
		c.Memmoryusage = containerinfo.Stats[len(containerinfo.Stats)-1].Memory.Usage
		nets := []*Net{}
		for index, inter := range containerinfo.Stats[len(containerinfo.Stats)-1].Network.Interfaces {
			net := &Net{}
			net.Name = inter.Name
			net.Rx = float64(inter.RxBytes-containerinfo.Stats[0].Network.Interfaces[index].RxBytes) * 1000 / float64(interval)
			net.Tx = float64(inter.TxBytes-containerinfo.Stats[0].Network.Interfaces[index].TxBytes) * 1000 / float64(interval)
			//net.Rx = float64(containerinfo.Stats[len(containerinfo.Stats)-1].Network.RxBytes-containerinfo.Stats[0].Network.RxBytes) * 1000 / float64(interval)
			//net.Tx = float64(containerinfo.Stats[len(containerinfo.Stats)-1].Network.TxBytes-containerinfo.Stats[0].Network.TxBytes) * 1000 / float64(interval)
			nets = append(nets, net)
		}
		c.NetworkInfo = nets
	}
	return err
}
func GetVersion(ip string) (DockerVer, KernelVer, OSVer string) {
	cmd := exec.Command("/bin/sh", "-c", "curl "+ip+" | grep \"Docker Version\"  | cut -d '<' -f4 | cut -d ' ' -f2")
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	DockerV := strings.TrimRight(out.String(), string(10))
	cmd2 := exec.Command("/bin/sh", "-c", "curl "+ip+" | grep \"Kernel Version\"  | cut -d '<' -f4 | cut -d ' ' -f2")
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	_ = cmd2.Run()
	KernelV := strings.TrimRight(out2.String(), string(10))
	cmd3 := exec.Command("/bin/sh", "-c", "curl "+ip+" | grep \"OS Version\"  | cut -d '<' -f4 | cut -d ' ' -f2")
	var out3 bytes.Buffer
	cmd3.Stdout = &out3
	_ = cmd3.Run()
	OSV := strings.TrimRight(out3.String(), string(10))
	fmt.Println("GetVersion", DockerV, KernelV, OSV)
	return DockerV, KernelV, OSV
}
