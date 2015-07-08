package main

import (
	"time"
)

type ContainerReference struct {
	Name      string   `json:"name"`
	Aliases   []string `json:"aliases,omitempty"`
	Namespace string   `json:"namespace,omitempty"`
}
type CpuSpec struct {
	Limit    uint64 `json:"limit"`
	MaxLimit uint64 `json:"max_limit"`
	Mask     string `json:"mask,omitempty"`
}
type MemorySpec struct {
	Limit       uint64 `json:"limit,omitempty"`
	Reservation uint64 `json:"reservation,omitempty"`
	SwapLimit   uint64 `json:"swap_limit,omitempty"`
}
type ContainerSpec struct {
	CreationTime  time.Time         `json:"creation_time,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	HasCpu        bool              `json:"has_cpu"`
	Cpu           CpuSpec           `json:"cpu,omitempty"`
	HasMemory     bool              `json:"has_memory"`
	Memory        MemorySpec        `json:"memory,omitempty"`
	HasNetwork    bool              `json:"has_network"`
	HasFilesystem bool              `json:"has_filesystem"`
	HasDiskIo     bool              `json:"has_diskio"`
}
type LoadStats struct {
	NrSleeping        uint64 `json:"nr_sleeping"`
	NrRunning         uint64 `json:"nr_running"`
	NrStopped         uint64 `json:"nr_stopped"`
	NrUninterruptible uint64 `json:"nr_uninterruptible"`
	NrIoWait          uint64 `json:"nr_io_wait"`
}
type CpuUsage struct {
	Total  uint64   `json:"total"`
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`
	User   uint64   `json:"user"`
	System uint64   `json:"system"`
}
type CpuStats struct {
	Usage       CpuUsage `json:"usage"`
	LoadAverage int32    `json:"load_average"`
}
type PerDiskStats struct {
	Major uint64            `json:"major"`
	Minor uint64            `json:"minor"`
	Stats map[string]uint64 `json:"stats"`
}
type DiskIoStats struct {
	IoServiceBytes []PerDiskStats `json:"io_service_bytes,omitempty"`
	IoServiced     []PerDiskStats `json:"io_serviced,omitempty"`
	IoQueued       []PerDiskStats `json:"io_queued,omitempty"`
	Sectors        []PerDiskStats `json:"sectors,omitempty"`
	IoServiceTime  []PerDiskStats `json:"io_service_time,omitempty"`
	IoWaitTime     []PerDiskStats `json:"io_wait_time,omitempty"`
	IoMerged       []PerDiskStats `json:"io_merged,omitempty"`
	IoTime         []PerDiskStats `json:"io_time,omitempty"`
}
type MemoryStatsMemoryData struct {
	Pgfault    uint64 `json:"pgfault"`
	Pgmajfault uint64 `json:"pgmajfault"`
}
type MemoryStats struct {
	Usage            uint64                `json:"usage"`
	WorkingSet       uint64                `json:"working_set"`
	ContainerData    MemoryStatsMemoryData `json:"container_data,omitempty"`
	HierarchicalData MemoryStatsMemoryData `json:"hierarchical_data,omitempty"`
}
type InterfaceStats struct {
	Name      string `json:"name"`
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}
type NetworkStats struct {
	InterfaceStats `json:",inline"`
	Interfaces     []InterfaceStats `json:"interfaces,omitempty"`
}
type FsStats struct {
	Device          string `json:"device,omitempty"`
	Limit           uint64 `json:"capacity"`
	Usage           uint64 `json:"usage"`
	Available       uint64 `json:"available"`
	ReadsCompleted  uint64 `json:"reads_completed"`
	ReadsMerged     uint64 `json:"reads_merged"`
	SectorsRead     uint64 `json:"sectors_read"`
	ReadTime        uint64 `json:"read_time"`
	WritesCompleted uint64 `json:"writes_completed"`
	WritesMerged    uint64 `json:"writes_merged"`
	SectorsWritten  uint64 `json:"sectors_written"`
	WriteTime       uint64 `json:"write_time"`
	IoInProgress    uint64 `json:"io_in_progress"`
	IoTime          uint64 `json:"io_time"`
	WeightedIoTime  uint64 `json:"weighted_io_time"`
}
type ContainerStats struct {
	Timestamp  time.Time    `json:"timestamp"`
	Cpu        CpuStats     `json:"cpu,omitempty"`
	DiskIo     DiskIoStats  `json:"diskio,omitempty"`
	Memory     MemoryStats  `json:"memory,omitempty"`
	Network    NetworkStats `json:"network,omitempty"`
	Filesystem []FsStats    `json:"filesystem,omitempty"`
	TaskStats  LoadStats    `json:"task_stats,omitempty"`
}
type ContainerInfo struct {
	ContainerReference
	Subcontainers []ContainerReference `json:"subcontainers,omitempty"`
	Spec          ContainerSpec        `json:"spec,omitempty"`
	Stats         []*ContainerStats    `json:"stats,omitempty"`
}
