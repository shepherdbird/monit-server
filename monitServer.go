package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
			if Cluster[ip].CpuMax < conf.Cpuusage {
				Cluster[ip].CpuMax = conf.Cpuusage
			}
			if Cluster[ip].DiskMax < conf.Diskusage {
				Cluster[ip].DiskMax = conf.Diskusage
			}
			if Cluster[ip].MemoryMax < conf.Memmoryusage {
				Cluster[ip].MemoryMax = conf.Memmoryusage
			}
			if Cluster[ip].RxMax < conf.NetworkInfo[0].Rx {
				Cluster[ip].RxMax = conf.NetworkInfo[0].Rx
			}
			if Cluster[ip].TxMax < conf.NetworkInfo[0].Tx {
				Cluster[ip].TxMax = conf.NetworkInfo[0].Tx
			}
			if len(Cluster[ip].Status) < 240 {
				Cluster[ip].Status = append(Cluster[ip].Status, conf)
			} else {
				Cluster[ip].Status[Cluster[ip].Point] = conf
			}
			Cluster[ip].Point = (Cluster[ip].Point + 1) % 240
		}
		time.Sleep(time.Second * 30)
	}

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
	resp, err := client.Get("http://10.10.103.250:8080/api/v1beta3/nodes/")
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &data)
	for _, item := range data.Items {
		Ips = append(Ips, item.Name)
		Cluster[item.Name] = NewNodestatus(item.Name)
	}
}
