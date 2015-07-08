package main

type FsInfo struct {
	Device   string `json:"device"`
	Capacity uint64 `json:"capacity"`
}
type DiskInfo struct {
	Name      string `json:"name"`
	Major     uint64 `json:"major"`
	Minor     uint64 `json:"minor"`
	Size      uint64 `json:"size"`
	Scheduler string `json:"scheduler"`
}
type NetInfo struct {
	Name       string `json:"name"`
	MacAddress string `json:"mac_address"`
	Speed      int64  `json:"speed"`
	Mtu        int64  `json:"mtu"`
}
type Core struct {
	Id      int     `json:"core_id"`
	Threads []int   `json:"thread_ids"`
	Caches  []Cache `json:"caches"`
}

type Cache struct {
	Size  uint64 `json:"size"`
	Type  string `json:"type"`
	Level int    `json:"level"`
}
type Node_machine struct {
	Id     int     `json:"node_id"`
	Memory uint64  `json:"memory"`
	Cores  []Core  `json:"cores"`
	Caches []Cache `json:"caches"`
}
type MachineInfo struct {
	NumCores       int                 `json:"num_cores"`
	CpuFrequency   uint64              `json:"cpu_frequency_khz"`
	MemoryCapacity int64               `json:"memory_capacity"`
	MachineID      string              `json:"machine_id"`
	SystemUUID     string              `json:"system_uuid"`
	BootID         string              `json:"boot_id"`
	Filesystems    []FsInfo            `json:"filesystems"`
	DiskMap        map[string]DiskInfo `json:"disk_map"`
	NetworkDevices []NetInfo           `json:"network_devices"`
	Topology       []Node_machine      `json:"topology"`
	CloudProvider  string              `json:"cloud_provider"`
	InstanceType   string              `json:"instance_type"`
}
