package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type ContainerNet struct {
	Name string
	Rx   float64
	Tx   float64
}
type ContainerConfig struct {
	Timestamp    int64
	Cpuusage     float64
	Memmoryusage uint64
	Diskusage    uint64
	NetworkInfo  ContainerNet
}
type ContainerNode struct {
	Creation_time int64
	Created_at    string
	Cpu_limit     uint64
	Memory_limit  uint64
	Fs_limit      uint64
	Spec          SpecialData
	Status        []*ContainerConfig
	Index         int
	Switch        bool
}

var (
	Container    map[string]*ContainerNode
	ConPointNums uint64
)

func GetCoreNumFromMask(mask string, numCores int) int {
	activeCores := GetActiveCores(mask)
	num := 0
	for i := 0; i < numCores; i++ {
		if activeCores[i] {
			num++
		}
	}
	return num
}

func GetActiveCores(mask string) map[int]bool {
	activeCores := make(map[int]bool)
	for _, corebits := range strings.Split(mask, ",") {
		cores := strings.Split(corebits, "-")
		if len(cores) == 1 {
			index, err := strconv.Atoi(cores[0])
			if err != nil {
				// Ignore malformed strings.
				continue
			}
			activeCores[index] = true
		} else if len(cores) == 2 {
			start, err := strconv.Atoi(cores[0])
			if err != nil {
				continue
			}
			end, err := strconv.Atoi(cores[1])
			if err != nil {
				continue
			}
			for i := start; i <= end; i++ {
				activeCores[i] = true
			}
		}
	}
	return activeCores
}

func ContainerIdInfo(ip string) {
	containerinfo := ContainerInfo{}
	client := &http.Client{}
	resp, err := client.Get("http://" + ip + ":4194/api/v1.3/containers/docker/")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		_ = json.Unmarshal(buff.Bytes(), &containerinfo)
		for _, subcontainers := range containerinfo.Subcontainers {
			_, v := Container[subcontainers.Name]
			if !v {
				Container[subcontainers.Name] = &ContainerNode{}
			}
			containerIdInfo := ContainerInfo{}
			resp1, err := client.Get("http://" + ip + ":4194/api/v1.3/containers" + subcontainers.Name)

			if err != nil {
				fmt.Println(err.Error())
			} else {
				buff1 := new(bytes.Buffer)
				buff1.ReadFrom(resp1.Body)
				_ = json.Unmarshal(buff1.Bytes(), &containerIdInfo)
				Container[subcontainers.Name].Creation_time = containerIdInfo.Spec.CreationTime.Unix()
				Container[subcontainers.Name].Cpu_limit = containerIdInfo.Spec.Cpu.Limit
				//Container[subcontainers.Name].Memory_limit = containerIdInfo.Spec.Memory.Limit
				Container[subcontainers.Name].Memory_limit = uint64(FCluster.Cluster[ip].Memorycapacity)
				Container[subcontainers.Name].Created_at = ip
				var filesystem uint64 = 0
				if containerIdInfo.Spec.HasFilesystem {
					Container[subcontainers.Name].Fs_limit = containerIdInfo.Stats[0].Filesystem[0].Limit
					for _, fs := range containerIdInfo.Stats[0].Filesystem {
						filesystem += fs.Usage
					}
				}
				if len(Container[subcontainers.Name].Status) < 120 {
					conf := &ContainerConfig{}
					Container[subcontainers.Name].Status = append(Container[subcontainers.Name].Status, conf)
				}
				//fmt.Println(len(containerIdInfo.Stats))
				if len(containerIdInfo.Stats) == 0 {
					fmt.Println("No container Info find.")
				} else {
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Diskusage = filesystem
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp = containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Timestamp.Unix()
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Memmoryusage = containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Memory.Usage
					interval := containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Timestamp.Sub(containerIdInfo.Stats[0].Timestamp)
					realContainerCoreNums := GetCoreNumFromMask(containerIdInfo.Spec.Cpu.Mask, Cluster[ip].Cpucores)
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Cpuusage = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Cpu.Usage.Total-containerIdInfo.Stats[0].Cpu.Usage.Total) / float64(interval) / float64(realContainerCoreNums)
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Name = containerIdInfo.Stats[0].Network.Name
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Rx = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Network.RxBytes-containerIdInfo.Stats[0].Network.RxBytes) * 1000000000 / float64(interval)
					Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Tx = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Network.TxBytes-containerIdInfo.Stats[0].Network.TxBytes) * 1000000000 / float64(interval)
				}
				if Container[subcontainers.Name].Spec.CpuMax < Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Cpuusage {
					Container[subcontainers.Name].Spec.CpuMax = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Cpuusage
					Container[subcontainers.Name].Spec.CpuMaxTimeStamp = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp
				}
				Container[subcontainers.Name].Spec.CpuAvg = (Container[subcontainers.Name].Spec.CpuAvg*float64(ConPointNums) + Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Cpuusage) / float64(ConPointNums+1)
				if Container[subcontainers.Name].Spec.DiskMax < Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Diskusage {
					Container[subcontainers.Name].Spec.DiskMax = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Diskusage
					Container[subcontainers.Name].Spec.DiskMaxTimeStamp = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp
				}
				Container[subcontainers.Name].Spec.DiskAvg = (Container[subcontainers.Name].Spec.DiskAvg*ConPointNums + Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Diskusage) / (ConPointNums + 1)
				if Container[subcontainers.Name].Spec.MemoryMax < Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Memmoryusage {
					Container[subcontainers.Name].Spec.MemoryMax = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Memmoryusage
					Container[subcontainers.Name].Spec.MemoryMaxTimeStamp = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp
				}
				Container[subcontainers.Name].Spec.MemoryAvg = (Container[subcontainers.Name].Spec.MemoryAvg*ConPointNums + Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Memmoryusage) / (ConPointNums + 1)
				if Container[subcontainers.Name].Spec.RxMax < Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Rx {
					Container[subcontainers.Name].Spec.RxMax = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Rx
					Container[subcontainers.Name].Spec.RxMaxTimeStamp = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp
				}
				Container[subcontainers.Name].Spec.RxAvg = (Container[subcontainers.Name].Spec.RxAvg*float64(ConPointNums) + Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Rx) / float64(ConPointNums+1)
				if Container[subcontainers.Name].Spec.TxMax < Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Tx {
					Container[subcontainers.Name].Spec.TxMax = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Tx
					Container[subcontainers.Name].Spec.TxMaxTimeStamp = Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].Timestamp
				}
				Container[subcontainers.Name].Spec.TxAvg = (Container[subcontainers.Name].Spec.TxAvg*float64(ConPointNums) + Container[subcontainers.Name].Status[Container[subcontainers.Name].Index].NetworkInfo.Tx) / float64(ConPointNums+1)
				Container[subcontainers.Name].Index = (Container[subcontainers.Name].Index + 1) % 120
				Container[subcontainers.Name].Switch = Switch
			}
		}
	}
}
