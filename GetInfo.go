package main

func (c *Config) GetMachineInfo(ip string) {
	machineinfo := MachineInfo{}
	client := &http.Client{}
	resp, err := client.Get("http://" + ip + ":4194/api/v1.3/machine/")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	buff := new(bytes.Buffer)
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &machineinfo)
	c.Cpucores = machineinfo.NumCores
	c.Cpufrequency = machineinfo.CpuFrequency * 1000
	c.Memorycapacity = machineinfo.MemoryCapacity
	var Filesystem uint64 = 0
	for _, fs := range machineinfo.Filesystems {
		Filesystem += fs.Capacity
	}
	c.Diskcapacity = Filesystem
}
func (c *Config) GetContainerInfo(ip string) {
	containerinfo := ContainerInfo{}
	resp, err = client.Get("http://" + ip + ":4194/api/v1.3/containers/")
	if err != nil {
		fmt.Println(err.Error())
	}
	defer resp.Body.Close()
	buff.ReadFrom(resp.Body)
	_ = json.Unmarshal(buff.Bytes(), &containerinfo)
	c.Timestamp = containerinfo.Spec.CreationTime
	c.Cpuusage = (containerinfo.Stats[len(containerinfo.Stats)-1].Cpu.Usage.Total - containerinfo.Stats[0].Cpu.Usage.Total) / (len(containerinfo.Stats) - 1)
	var Filesystem uint64 = 0
	for _, fs := range containerinfo.Stats[0].Filesystem {
		Filesystem += fs.Usage
	}
	c.Diskusage = Filesystem
	c.Memmoryusage = containerinfo.Stats[0].Memory.Usage
}
