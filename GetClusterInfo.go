package main

import (
	//"time"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
	CpuMax             uint64
	CpuMaxTimeStamp    int64
	MemoryMax          uint64
	MemoryMaxTimeStamp int64
	DiskMax            uint64
	DiskMaxTimeStamp   int64
	RxMax              float64
	RxMaxTimeStamp     int64
	TxMax              float64
	TxMaxTimeStamp     int64
}
type Nodestatus struct {
	//Ip             string
	Cpucores       int
	Cpufrequency   uint64
	Memorycapacity int64
	Diskcapacity   uint64
	Spec           SpecialData
	Status         []*Config
	Index          int
}
type FinalCluster struct {
	Status  string
	Cluster map[string]*Nodestatus
}

var (
	Ips      []string
	Cluster  map[string]*Nodestatus
	FCluster FinalCluster
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
func (c *Config) GetMachineInfo(ip string) {
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
}
func (c *Config) GetContainerInfo(ip string) {
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
}
