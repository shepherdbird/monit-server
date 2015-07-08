package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

var (
	Ips     []string
	Cluster map[string]*Nodestatus
)

func status(w http.ResponseWriter, req *http.Request) {
	Jdata, _ := json.Marshal(Cluster)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(Jdata))

}
func getStatus() {
	for {
		for _, ip := range Ips {
			conf := &Config{}
			conf.GetMachineInfo(ip)
			conf.GetContainerInfo(ip)
			if Cluster[ip].Spec.CpuMax < conf.Cpuusage {
				Cluster[ip].Spec.CpuMax = conf.Cpuusage
				Cluster[ip].Spec.CpuMaxTimeStamp = conf.Timestamp
			}
			if Cluster[ip].Spec.DiskMax < conf.Diskusage {
				Cluster[ip].Spec.DiskMax = conf.Diskusage
				Cluster[ip].Spec.DiskMaxTimeStamp = conf.Timestamp
			}
			if Cluster[ip].Spec.MemoryMax < conf.Memmoryusage {
				Cluster[ip].Spec.MemoryMax = conf.Memmoryusage
				Cluster[ip].Spec.MemoryMaxTimeStamp = conf.Timestamp
			}
			if Cluster[ip].Spec.RxMax < conf.NetworkInfo[0].Rx {
				Cluster[ip].Spec.RxMax = conf.NetworkInfo[0].Rx
				Cluster[ip].Spec.RxMaxTimeStamp = conf.Timestamp
			}
			if Cluster[ip].Spec.TxMax < conf.NetworkInfo[0].Tx {
				Cluster[ip].Spec.TxMax = conf.NetworkInfo[0].Tx
				Cluster[ip].Spec.TxMaxTimeStamp = conf.Timestamp
			}
			if len(Cluster[ip].Status) < 240 {
				Cluster[ip].Status = append(Cluster[ip].Status, conf)
			} else {
				Cluster[ip].Status[Cluster[ip].Index] = conf
			}
			Cluster[ip].Index = (Cluster[ip].Index + 1) % 240
		}
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
func main() {
	go getStatus()
	http.HandleFunc("/api/status", status)
	err := http.ListenAndServe("0.0.0.0:50000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
func init() {
	Ips = []string{}
	Cluster = map[string]*Nodestatus{}
	data := NodeList{}
	client := &http.Client{}
	//resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
	resp, err := client.Get("http://" + GetLocalIp() + ":8080/api/v1beta3/nodes/")
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &data)
	for _, item := range data.Items {
		Ips = append(Ips, item.Name)
		Cluster[item.Name] = &Nodestatus{} //NewNodestatus(item.Name)
	}
}
