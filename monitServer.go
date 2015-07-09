package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//"log"
	"github.com/coreos/go-etcd/etcd"
	"net"
	"net/http"
	"os"
	"time"
)

var Switch bool = true
var EtcdClient *etcd.Client
var Tocken = "123456"

func ClusterStatus(w http.ResponseWriter, req *http.Request) {
	FCluster.Cluster = Cluster
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		Jdata, _ := json.Marshal(FCluster)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(Jdata))
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}
func ContainerStatus(w http.ResponseWriter, req *http.Request) {
	containerId := req.Header.Get("container")
	if req.Header.Get("token") == "qwertyuiopasdfghjklzxcvbnm1234567890" {
		k, v := Container["/docker/"+containerId]
		if v {
			Jdata, _ := json.Marshal(k)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(Jdata))
		} else {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("container id do not exist"))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
	}

}
func GetClusterStatus() {
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
			if len(Cluster[ip].Status) < 120 {
				Cluster[ip].Status = append(Cluster[ip].Status, conf)
			} else {
				Cluster[ip].Status[Cluster[ip].Index] = conf
			}
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

func main() {
	go GetClusterStatus()
	go GetContainerStatus()
	http.HandleFunc("/api/cluster/status", ClusterStatus)
	http.HandleFunc("/api/container/status", ContainerStatus)
	EtcdClient = etcd.NewClient([]string{"http://127.0.0.1:2379"})
	http.HandleFunc("/create", Create)
	http.HandleFunc("/get", Get)
	http.HandleFunc("/delete", Delete)
	//err := http.ListenAndServeTLS("0.0.0.0:50000", "cert.pem", "key.pem", nil)
	err := http.ListenAndServe("0.0.0.0:50000", nil)
	if err != nil {
		fmt.Println(err.Error())
	}
}
func init() {
	Ips = []string{}
	Cluster = map[string]*Nodestatus{}
	Container = map[string]*ContainerNode{}
	FCluster.Status = "Ready"
	FCluster.Cluster = Cluster
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
			if item.Status.Conditions[0].Type != "Ready" {
				FCluster.Status = "Not Ready"
			}
		}
	}
}
