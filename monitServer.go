package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"log"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type OverviewNode struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

var (
	Switch     bool = true
	EtcdClient *etcd.Client
	Tocken         = "qwertyuiopasdfghjklzxcvbnm1234567890"
	RxSecond   int = 0
	TxSecond   int = 0
)

func ClusterStatus(w http.ResponseWriter, req *http.Request) {
	FCluster.Cluster = Cluster
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		if req.Header.Get("flag") == "1" {
			Jdata, _ := json.Marshal(FCluster)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(Jdata))
		} else {
			LCluster := FinalCluster{}
			LCluster.Status = FCluster.Status
			LCluster.MasterRxAvg = FCluster.MasterRxAvg
			LCluster.MasterRxMax = FCluster.MasterRxMax
			LCluster.MasterRxMaxStamp = FCluster.MasterRxMaxStamp
			LCluster.MasterTxAvg = FCluster.MasterTxAvg
			LCluster.MasterTxMax = FCluster.MasterTxMax
			LCluster.MasterTxMaxStamp = FCluster.MasterTxMaxStamp
			nets := &ClusterNetwork{}
			nets.TimeStamp = time.Now().Unix()
			nets.Rx, nets.Tx = GetClusterNetworkStatus()
			LCluster.Network = append(LCluster.Network, nets)
			LLcluster := map[string]*Nodestatus{}
			for _, ip := range Ips {
				LNode := &Nodestatus{}
				LNode.Cpucores = FCluster.Cluster[ip].Cpucores
				LNode.Cpufrequency = FCluster.Cluster[ip].Cpufrequency
				LNode.Diskcapacity = FCluster.Cluster[ip].Diskcapacity
				LNode.DockerVersion = FCluster.Cluster[ip].DockerVersion
				LNode.KernelVersion = FCluster.Cluster[ip].KernelVersion
				LNode.Memorycapacity = FCluster.Cluster[ip].Memorycapacity
				LNode.OSVersion = FCluster.Cluster[ip].OSVersion
				LNode.Spec = FCluster.Cluster[ip].Spec
				LNode.NetSpec = FCluster.Cluster[ip].NetSpec
				conf := &Config{}
				err := conf.GetContainerInfo(ip)
				if err != nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Failed to get machine info."))
				} else {
					LNode.Status = append(LNode.Status, conf)
					LLcluster[ip] = LNode
				}
			}
			LCluster.Cluster = LLcluster
			Jdata, _ := json.Marshal(LCluster)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(Jdata))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}
func ContainerStatus(w http.ResponseWriter, req *http.Request) {
	containerId := req.Header.Get("container")
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		k, v := Container["/docker/"+containerId]
		if v {
			if req.Header.Get("flag") == "1" {
				Jdata, _ := json.Marshal(k)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(Jdata))
			} else {
				LContainer := &ContainerNode{}
				LContainer.Cpu_limit = k.Cpu_limit
				LContainer.Created_at = k.Created_at
				LContainer.Creation_time = k.Creation_time
				LContainer.Fs_limit = k.Fs_limit
				LContainer.Memory_limit = k.Memory_limit
				LContainer.Spec = k.Spec
				client := &http.Client{}
				containerIdInfo := ContainerInfo{}
				resp1, err := client.Get("http://" + k.Created_at + ":4194/api/v1.3/containers" + "/docker/" + containerId)
				if err != nil {
					fmt.Println(err.Error())
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
				} else {
					buff1 := new(bytes.Buffer)
					buff1.ReadFrom(resp1.Body)
					_ = json.Unmarshal(buff1.Bytes(), &containerIdInfo)
					conf := &ContainerConfig{}
					LContainer.Status = append(LContainer.Status, conf)
					var filesystem uint64 = 0
					if containerIdInfo.Spec.HasFilesystem {
						for _, fs := range containerIdInfo.Stats[0].Filesystem {
							filesystem += fs.Usage
						}
					}
					//fmt.Println(len(containerIdInfo.Stats))
					if len(containerIdInfo.Stats) == 0 {
						fmt.Println("No container Info find.")
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("No container Info find."))
					} else {
						LContainer.Status[0].Diskusage = filesystem
						LContainer.Status[0].Timestamp = containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Timestamp.Unix()
						LContainer.Status[0].Memmoryusage = containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Memory.Usage
						interval := containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Timestamp.Sub(containerIdInfo.Stats[0].Timestamp)
						realContainerCoreNums := GetCoreNumFromMask(containerIdInfo.Spec.Cpu.Mask, Cluster[k.Created_at].Cpucores)
						LContainer.Status[0].Cpuusage = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Cpu.Usage.Total-containerIdInfo.Stats[0].Cpu.Usage.Total) / float64(interval) / float64(realContainerCoreNums)
						LContainer.Status[0].NetworkInfo.Name = containerIdInfo.Stats[0].Network.Name
						LContainer.Status[0].NetworkInfo.Rx = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Network.RxBytes-containerIdInfo.Stats[0].Network.RxBytes) * 1000000000 / float64(interval)
						LContainer.Status[0].NetworkInfo.Tx = float64(containerIdInfo.Stats[len(containerIdInfo.Stats)-1].Network.TxBytes-containerIdInfo.Stats[0].Network.TxBytes) * 1000000000 / float64(interval)
						Jdata, _ := json.Marshal(LContainer)
						w.Header().Set("Content-Type", "application/json; charset=utf-8")
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(Jdata))
					}
				}
			}
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("container id do not exist"))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}
func GetNodes(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		data := NodeList{}
		client := &http.Client{}
		//resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
		resp, err := client.Get("http://" + GetLocalIp() + ":8080/api/v1beta3/nodes/")
		if err != nil {
			fmt.Printf(err.Error())
			w.WriteHeader(http.StatusForbidden)
		} else {
			defer resp.Body.Close()
			buff := new(bytes.Buffer)
			buff.ReadFrom(resp.Body)
			_ = json.Unmarshal(buff.Bytes(), &data)
			OverNodes := []OverviewNode{}
			for _, item := range data.Items {
				//Ips = append(Ips, item.Name)
				Lin := OverviewNode{}
				Lin.Ip = item.Name
				Lin.Status = "Ready"
				//Cluster[item.Name] = &Nodestatus{} //NewNodestatus(item.Name)
				//fmt.Println("item:", item.Name)
				if item.Status.Conditions[0].Type != "Ready" {
					Lin.Status = "Not Ready"
				}
				OverNodes = append(OverNodes, Lin)
			}
			Jdata, _ := json.Marshal(OverNodes)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(Jdata))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}
func ApiNodeStatus(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		ip := req.Header.Get("node")
		if req.Header.Get("flag") == "1" {
			Jdata, _ := json.Marshal(FCluster.Cluster[ip])
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(Jdata))
		} else {
			LNode := &Nodestatus{}
			LNode.Cpucores = FCluster.Cluster[ip].Cpucores
			LNode.Cpufrequency = FCluster.Cluster[ip].Cpufrequency
			LNode.Diskcapacity = FCluster.Cluster[ip].Diskcapacity
			LNode.DockerVersion = FCluster.Cluster[ip].DockerVersion
			LNode.KernelVersion = FCluster.Cluster[ip].KernelVersion
			LNode.Memorycapacity = FCluster.Cluster[ip].Memorycapacity
			LNode.OSVersion = FCluster.Cluster[ip].OSVersion
			LNode.Spec = FCluster.Cluster[ip].Spec
			LNode.NetSpec = FCluster.Cluster[ip].NetSpec
			conf := &Config{}
			err := conf.GetContainerInfo(ip)
			if err != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to get machine info."))
			} else {
				LNode.Status = append(LNode.Status, conf)
				Jdata, _ := json.Marshal(LNode)
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(Jdata))
			}
		}

	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}
func excute(command string) string {
	cmd := exec.Command("/bin/sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	_ = cmd.Run()
	return out.String()
}
func GetClusterNetworkStatus() (Rx float64, Tx float64) {
	RxOld := excute("iptables -nvxL INPUT | grep tcp | awk 'BEGIN{count=0;}{if(NR==1){count+=$2;}else{count-=$2;}}END{print count;}'")
	RxFirst, _ := strconv.Atoi(strings.TrimRight(RxOld, string(10)))
	TxOld := excute("iptables -nvxL OUTPUT | grep tcp | awk 'BEGIN{count=0;}{if(NR==1){count+=$2;}else{count-=$2;}}END{print count;}'")
	TxFirst, _ := strconv.Atoi(strings.TrimRight(TxOld, string(10)))
	if len(FCluster.Network) == 0 {
		time.Sleep(time.Second * 1)
		RxNew := excute("iptables -nvxL INPUT | grep tcp | awk 'BEGIN{count=0;}{if(NR==1){count+=$2;}else{count-=$2;}}END{print count;}'")
		RxSecond, _ := strconv.Atoi(strings.TrimRight(RxNew, string(10)))
		TxNew := excute("iptables -nvxL OUTPUT | grep tcp | awk 'BEGIN{count=0;}{if(NR==1){count+=$2;}else{count-=$2;}}END{print count;}'")
		TxSecond, _ := strconv.Atoi(strings.TrimRight(TxNew, string(10)))
		fmt.Println(int64(RxSecond), int64(TxSecond))
		return float64(RxSecond - RxFirst), float64(TxSecond - TxFirst)
	} else {
		RxTemp, TxTemp := float64(RxFirst-RxSecond)/30.0, float64(TxFirst-TxSecond)/30.0
		RxSecond = RxFirst
		TxSecond = TxFirst
		return RxTemp, TxTemp
	}
}

func MonitDockerDaemon() {
	for {
		_, err := EtcdClient.Get("/", false, false)
		if err != nil {
			fmt.Printf("Docker Daemon is crashed. Send a email. %v\n", err)
			ip := GetLocalIp()
			content := "<html><body>Machine IP:" + ip + "</body></html>"
			SendEmail(
				"smtp.126.com",
				25,
				"wonderflow@126.com",
				"zjuvlis123456",
				[]string{"544028616@qq.com"},
				"Docker Daemon crashed",
				content)
			time.Sleep(time.Hour)
		}
		time.Sleep(time.Second * 30)
	}
}

func GetClusterStatus() {
	Point = 0
	for {
		for _, ip := range Ips {
			conf := &Config{}
			err := conf.GetMachineInfo(ip)
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
			err = conf.GetContainerInfo(ip)
			if err != nil {
				time.Sleep(time.Second * 10)
				continue
			}
			if Cluster[ip].Spec.CpuMax < conf.Cpuusage {
				Cluster[ip].Spec.CpuMax = conf.Cpuusage
				Cluster[ip].Spec.CpuMaxTimeStamp = conf.Timestamp
			}
			Cluster[ip].Spec.CpuAvg = (Cluster[ip].Spec.CpuAvg*float64(PointNums) + conf.Cpuusage) / float64(PointNums+1)
			if Cluster[ip].Spec.DiskMax < conf.Diskusage {
				Cluster[ip].Spec.DiskMax = conf.Diskusage
				Cluster[ip].Spec.DiskMaxTimeStamp = conf.Timestamp
			}
			Cluster[ip].Spec.DiskAvg = (Cluster[ip].Spec.DiskAvg*PointNums + conf.Diskusage) / (PointNums + 1)
			if Cluster[ip].Spec.MemoryMax < conf.Memmoryusage {
				Cluster[ip].Spec.MemoryMax = conf.Memmoryusage
				Cluster[ip].Spec.MemoryMaxTimeStamp = conf.Timestamp
			}
			Cluster[ip].Spec.MemoryAvg = (Cluster[ip].Spec.MemoryAvg*PointNums + conf.Memmoryusage) / (PointNums + 1)
			if Cluster[ip].Spec.RxMax < conf.NetworkInfo[0].Rx {
				Cluster[ip].Spec.RxMax = conf.NetworkInfo[0].Rx
				Cluster[ip].Spec.RxMaxTimeStamp = conf.Timestamp
			}
			Cluster[ip].Spec.RxAvg = (Cluster[ip].Spec.RxAvg*float64(PointNums) + float64(conf.NetworkInfo[0].Rx)) / float64(PointNums+1)
			if Cluster[ip].Spec.TxMax < conf.NetworkInfo[0].Tx {
				Cluster[ip].Spec.TxMax = conf.NetworkInfo[0].Tx
				Cluster[ip].Spec.TxMaxTimeStamp = conf.Timestamp
			}
			Cluster[ip].Spec.TxAvg = (Cluster[ip].Spec.TxAvg*float64(PointNums) + float64(conf.NetworkInfo[0].Tx)) / float64(PointNums+1)
			if len(Cluster[ip].Status) < 120 {
				Cluster[ip].Status = append(Cluster[ip].Status, conf)
			} else {
				Cluster[ip].Status[Cluster[ip].Index] = conf
			}
			//node network max and average
			if len(Cluster[ip].NetSpec) == 0 {
				for _, s := range conf.NetworkInfo {
					Lin := &NetworkSpec{}
					Lin.Name = s.Name
					Lin.RxMax = s.Rx
					Lin.RxMaxTimeStamp = conf.Timestamp
					Lin.RxAvg = s.Rx
					Lin.TxMaxTimeStamp = conf.Timestamp
					Lin.TxMax, Lin.TxAvg = s.Tx, s.Tx
					Cluster[ip].NetSpec = append(Cluster[ip].NetSpec, Lin)
				}
			} else {
				for k, v := range conf.NetworkInfo {
					if Cluster[ip].NetSpec[k].RxMax < v.Rx {
						Cluster[ip].NetSpec[k].RxMax = v.Rx
						Cluster[ip].NetSpec[k].RxMaxTimeStamp = conf.Timestamp
					}
					Cluster[ip].NetSpec[k].RxAvg = float64(Cluster[ip].NetSpec[k].RxAvg*float64(PointNums)+float64(v.Rx)) / float64(PointNums+1)
					if Cluster[ip].NetSpec[k].TxMax < v.Tx {
						Cluster[ip].NetSpec[k].TxMax = v.Tx
						Cluster[ip].NetSpec[k].TxMaxTimeStamp = conf.Timestamp
					}
					Cluster[ip].NetSpec[k].TxAvg = float64(Cluster[ip].NetSpec[k].TxAvg*float64(PointNums)+float64(v.Tx)) / float64(PointNums+1)
				}
			}
			Point = Cluster[ip].Index
			Cluster[ip].Index = (Cluster[ip].Index + 1) % 120
		}
		data := NodeList{}
		client := &http.Client{}
		//resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
		resp, err := client.Get("http://" + GetLocalIp() + ":8080/api/v1beta3/nodes/")
		if err != nil {
			fmt.Printf(err.Error())
		} else {
			FCluster.Status = "Ready"
			buff := new(bytes.Buffer)
			buff.ReadFrom(resp.Body)
			_ = json.Unmarshal(buff.Bytes(), &data)
			for _, item := range data.Items {
				if item.Status.Conditions[0].Type != "Ready" {
					FCluster.Status = "Not Ready"
				}
			}
		}
		nets := &ClusterNetwork{}
		nets.TimeStamp = time.Now().Unix()
		nets.Rx, nets.Tx = GetClusterNetworkStatus()
		fmt.Println(nets.Rx, nets.Tx)
		if len(FCluster.Network) < 120 {
			FCluster.Network = append(FCluster.Network, nets)
		} else {
			FCluster.Network[Point] = nets
		}
		if FCluster.MasterRxMax < nets.Rx {
			FCluster.MasterRxMax = nets.Rx
			FCluster.MasterRxMaxStamp = nets.TimeStamp
		}
		if FCluster.MasterTxMax < nets.Tx {
			FCluster.MasterTxMax = nets.Tx
			FCluster.MasterTxMaxStamp = nets.TimeStamp
		}
		FCluster.MasterRxAvg = (FCluster.MasterRxAvg*float64(PointNums) + nets.Rx) / float64(PointNums+1)
		FCluster.MasterTxAvg = (FCluster.MasterTxAvg*float64(PointNums) + nets.Tx) / float64(PointNums+1)
		PointNums += 1
		time.Sleep(time.Second * 30)
	}

}
func GetContainerStatus() {
	for {
		Switch = !Switch
		for _, ip := range Ips {
			ContainerIdInfo(ip)
		}
		for k, v := range Container {
			if v.Switch != Switch {
				delete(Container, k)
			}
		}
		ConPointNums += 1
		time.Sleep(time.Second * 30)
	}
}
func GetLocalIp() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		fmt.Printf("Get Local IP error: %v\n", err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

/*work for etcd*/
func Create(w http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("Authorization")
	if token != Tocken {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+"Authorization Not Set"+`"}`, 406)
		return
	}
	key := req.FormValue("key")
	value := req.FormValue("value")
	if key == "" || value == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+fmt.Sprintf("key = %s ,value = %s", key, value)+`"}`, 406)
		return
	}
	_, err := EtcdClient.Get(key, false, false)
	if err != nil {
		_, err = EtcdClient.Create(key, value, 0)
	} else {
		_, err = EtcdClient.Update(key, value, 0)
	}
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+err.Error()+`"}`, 406)
		return
	}
	w.WriteHeader(200)

	w.Write([]byte(`{"Message":"` + "Success" + `"}`))
}
func Update(w http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("Authorization")
	if token != Tocken {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+"Authorization Not Set"+`"}`, 406)
		return
	}
	key := req.FormValue("key")
	value := req.FormValue("value")
	if key == "" || value == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+fmt.Sprintf("key = %s ,value = %s", key, value)+`"}`, 406)
		return
	}
	_, err := EtcdClient.Update(key, value, 0)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+err.Error()+`"}`, 406)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"Message":"` + "Success" + `"}`))
}
func Get(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	token := req.Header.Get("Authorization")
	if token != Tocken {
		//w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+"Authorization Not Set"+`"}`, 406)
		return
	}
	req.ParseForm()
	key := req.Form.Get("key")
	if key == "" {
		//w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+fmt.Sprintf("key = %s", key)+`"}`, 406)
		return
	}
	resp, err := EtcdClient.Get(key, false, false)
	if err != nil {
		//w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+err.Error()+`"}`, 406)
		return
	}
	tt, _ := json.Marshal(resp)
	fmt.Println(string(tt))
	w.WriteHeader(200)
	//w.Header().Set("Content-Type", "application/json")
	w.Write(tt)
}
func Delete(w http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("Authorization")
	if token != Tocken {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+"Authorization Not Set"+`"}`, 406)
		return
	}
	//req.ParseForm()
	//body, _ := ioutil.ReadAll(req.Body)
	//fmt.Println(string(body))
	//req.ParseForm()
	//key := req.Form.Get("key")
	key := req.FormValue("key")
	//key := req.FormValue("key")
	if key == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+fmt.Sprintf("key = %s", key)+`"}`, 406)
		return
	}
	_, err := EtcdClient.Delete(key, false)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+err.Error()+`"}`, 406)
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(`{"Message":"` + "Success" + `"}`))
}
func RMi(w http.ResponseWriter, req *http.Request) {
	token := req.Header.Get("Authorization")
	if token != Tocken {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+"Authorization Not Set"+`"}`, 406)
		return
	}
	image := req.FormValue("imagesname")
	//value := req.FormValue("value")
	client := &http.Client{}
	//resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
	resp, err := client.Get("http://" + GetLocalIp() + ":8080/api/v1beta3/nodes/")
	defer resp.Body.Close()
	if err != nil {
		//w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"errorMessage":"`+err.Error()+`"}`, 406)
		return
	}
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	data := NodeList{}
	_ = json.Unmarshal(buff.Bytes(), &data)

	chs := []chan int{}
	for _, item := range data.Items {
		ch := make(chan int)
		chs = append(chs, ch)
		nodeip := item.Name
		fmt.Println(nodeip)
		go func(nodeip string, ch chan int) {
			defer close(ch)
			endpoint := "http://" + nodeip + ":2376"
			dclient, err := docker.NewClient(endpoint)
			opts := docker.RemoveImageOptions{true, false}
			err = dclient.RemoveImageExtended(image, opts)
			if err != nil {
				for err != nil && strings.Contains(err.Error(), "409") {
					fmt.Println(err)
					data := strings.Split(err.Error(), " ")
					containerid := data[12]
					fmt.Println(containerid)
					rmc := docker.RemoveContainerOptions{ID: containerid, Force: true}
					dclient.RemoveContainer(rmc)
					err = dclient.RemoveImageExtended(image, opts)
					if err != nil {
						fmt.Println(err)
					}
				}
				if err == nil {
					fmt.Println(nodeip + "delete ok")
				}
			} else {
				fmt.Println(nodeip + "delete ok")
			}
		}(nodeip, ch)
	}
	for _, ch := range chs {
		<-ch
	}
	w.WriteHeader(200)
	//w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("delete success"))
}
func main() {
	go GetClusterStatus()
	go GetContainerStatus()

	excute("iptables -F INPUT")
	excute("iptables -F OUTPUT")
	////Rx
	excute("iptables -I INPUT -s 10.0.0.0/8 -p tcp -m multiport --dport 50000,8080,8081,2376,80")
	excute("iptables -I INPUT -s 172.16.0.0/16 -p tcp -m multiport --dport 50000,8080,8081,2376,80")
	excute("iptables -I INPUT -s 192.168.0.0/16 -p tcp -m multiport --dport 50000,8080,8081,2376,80")
	excute("iptables -I INPUT -p tcp -m multiport --dport 50000,8080,8081,2376,80")
	//Tx
	excute("iptables -I OUTPUT -d 10.0.0.0/8 -p tcp -m multiport --sport 50000,8080,8081,2376,80")
	excute("iptables -I OUTPUT -d 172.16.0.0/16 -p tcp -m multiport --sport 50000,8080,8081,2376,80")
	excute("iptables -I OUTPUT -d 192.168.0.0/16 -p tcp -m multiport --sport 50000,8080,8081,2376,80")
	excute("iptables -I OUTPUT -p tcp -m multiport --sport 50000,8080,8081,2376,80")
	http.HandleFunc("/api/cluster/status", ClusterStatus)
	http.HandleFunc("/api/container/status", ContainerStatus)
	http.HandleFunc("/api/cluster/nodes", GetNodes)
	http.HandleFunc("/api/node/status", ApiNodeStatus)
	EtcdClient = etcd.NewClient([]string{"http://127.0.0.1:4001"})

	go MonitDockerDaemon()

	http.HandleFunc("/create", Create)
	http.HandleFunc("/get", Get)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/rmi", RMi)
	err := http.ListenAndServeTLS("0.0.0.0:50000", "cert.pem", "key.pem", nil)
	//err := http.ListenAndServe("0.0.0.0:50000", nil)
	if err != nil {
		fmt.Println(err.Error())
	}

}
func init() {
	Ips = []string{}
	PointNums = 0
	ConPointNums = 0
	Cluster = map[string]*Nodestatus{}
	Container = map[string]*ContainerNode{}
	FCluster.Status = "Ready"
	FCluster.Network = []*ClusterNetwork{}
	FCluster.Cluster = Cluster
	FCluster.MasterRxMax = 0
	FCluster.MasterTxMax = 0
	data := NodeList{}
	client := &http.Client{}
	//resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
	resp, err := client.Get("http://" + GetLocalIp() + ":8080/api/v1beta3/nodes/")
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		defer resp.Body.Close()
		buff := new(bytes.Buffer)
		buff.ReadFrom(resp.Body)
		_ = json.Unmarshal(buff.Bytes(), &data)
		for _, item := range data.Items {
			Ips = append(Ips, item.Name)
			Cluster[item.Name] = &Nodestatus{} //NewNodestatus(item.Name)
			fmt.Println("item:", item.Name)
			Cluster[item.Name].DockerVersion, Cluster[item.Name].KernelVersion, Cluster[item.Name].OSVersion = GetVersion(item.Name)
			if item.Status.Conditions[0].Type != "Ready" {
				FCluster.Status = "Not Ready"
			}
		}
	}
}
