package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Net struct {
	Name string
	Rx   float64
	Tx   float64
}
type Config struct {
	Timestamp      time.Time
	Cpuusage       uint64
	Cpufrequency   uint64
	Cpucores       int
	Memmoryusage   uint64
	Memorycapacity uint64
	Diskusage      uint64
	Diskcapacity   uint64
	NetworkInfo    []Net
}
type NodeStatus struct {
	Ip     string
	Status []Config
}

var (
	Ips     []string
	Cluster []NodeStatus
)

func status(w http.ResponseWriter, req *http.Request) {
	fmt.Println("%v", Ips)
	Jdata, _ := json.Marshal(Ips)
	w.Write([]byte(Jdata))
}
func getStatus() {

	time.Sleep(time.Second * 30)

}
func main() {
	getStatus()
	http.HandleFunc("/api/status", status)
	err := http.ListenAndServe("127.0.0.1:50000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
	}
}
func init() {
	Ips = []string{}
	Cluster = []NodeStatus{}
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
		fmt.Println(item.Name)
	}
}
