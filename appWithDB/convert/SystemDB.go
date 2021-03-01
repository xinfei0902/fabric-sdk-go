package convert

import (
	"github.com/jinzhu/gorm"
)

// TimesStat contains the amounts of time the CPU has spent performing different
// kinds of work. Time units are in USER_HZ or Jiffies (typically hundredths of
// a second). It is based on linux /proc/stat file.
type TimesStat struct {
	gorm.Model              `json:"-"`
	PeerSystemInformationID uint `gorm:"index" json:"-"`

	CPU       string  `gorm:"varchar(255)" json:"cpu"`
	User      float64 `json:"user"`
	System    float64 `json:"system"`
	Idle      float64 `json:"idle"`
	Nice      float64 `json:"nice"`
	Iowait    float64 `json:"iowait"`
	Irq       float64 `json:"irq"`
	Softirq   float64 `json:"softirq"`
	Steal     float64 `json:"steal"`
	Guest     float64 `json:"guest"`
	GuestNice float64 `json:"guestNice"`
	Stolen    float64 `json:"stolen"`
}

// Memory usage statistics. Total, Available and Used contain numbers of bytes
// for human consumption.
//
// The other fields in this struct contain kernel specific values.
type VirtualMemoryStat struct {
	gorm.Model              `json:"-"`
	PeerSystemInformationID uint `gorm:"index" json:"-"`

	// Total amount of RAM on this system
	Total uint64 `json:"total"`

	// RAM available for programs to allocate
	//
	// This value is computed from the kernel specific values.
	Available uint64 `json:"available"`

	// RAM used by programs
	//
	// This value is computed from the kernel specific values.
	Used uint64 `json:"used"`

	// Percentage of RAM used by programs
	//
	// This value is computed from the kernel specific values.
	UsedPercent float64 `json:"usedPercent"`

	// This is the kernel's notion of free memory; RAM chips whose bits nobody
	// cares about the value of right now. For a human consumable number,
	// Available is what you really want.
	Free uint64 `json:"free"`

	// OS X / BSD specific numbers:
	// http://www.macyourself.com/2010/02/17/what-is-free-wired-active-and-inactive-system-memory-ram/
	Active   uint64 `json:"active"`
	Inactive uint64 `json:"inactive"`
	Wired    uint64 `json:"wired"`

	// Linux specific numbers
	// https://www.centos.org/docs/5/html/5.1/Deployment_Guide/s2-proc-meminfo.html
	// https://www.kernel.org/doc/Documentation/filesystems/proc.txt
	// https://www.kernel.org/doc/Documentation/vm/overcommit-accounting
	Buffers        uint64 `json:"buffers"`
	Cached         uint64 `json:"cached"`
	Writeback      uint64 `json:"writeback"`
	Dirty          uint64 `json:"dirty"`
	WritebackTmp   uint64 `json:"writebacktmp"`
	Shared         uint64 `json:"shared"`
	Slab           uint64 `json:"slab"`
	PageTables     uint64 `json:"pagetables"`
	SwapCached     uint64 `json:"swapcached"`
	CommitLimit    uint64 `json:"commitlimit"`
	CommittedAS    uint64 `json:"committedas"`
	HighTotal      uint64 `json:"hightotal"`
	HighFree       uint64 `json:"highfree"`
	LowTotal       uint64 `json:"lowtotal"`
	LowFree        uint64 `json:"lowfree"`
	SwapTotal      uint64 `json:"swaptotal"`
	SwapFree       uint64 `json:"swapfree"`
	Mapped         uint64 `json:"mapped"`
	VMallocTotal   uint64 `json:"vmalloctotal"`
	VMallocUsed    uint64 `json:"vmallocused"`
	VMallocChunk   uint64 `json:"vmallocchunk"`
	HugePagesTotal uint64 `json:"hugepagestotal"`
	HugePagesFree  uint64 `json:"hugepagesfree"`
	HugePageSize   uint64 `json:"hugepagesize"`
}

type SwapMemoryStat struct {
	gorm.Model              `json:"-"`
	PeerSystemInformationID uint `gorm:"index" json:"-"`

	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"usedPercent"`
	Sin         uint64  `json:"sin"`
	Sout        uint64  `json:"sout"`
}

type UsageStat struct {
	gorm.Model              `json:"-"`
	PeerSystemInformationID uint `gorm:"index" json:"-"`

	Path              string  `gorm:"varchar(255)" json:"path"`
	Fstype            string  `gorm:"varchar(255)" json:"fstype"`
	Total             uint64  `json:"total"`
	Free              uint64  `json:"free"`
	Used              uint64  `json:"used"`
	UsedPercent       float64 `json:"usedPercent"`
	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

// ChainType for report
type ChainType struct {
	gorm.Model              `json:"-"`
	PeerSystemInformationID uint `gorm:"index" json:"-"`

	Name    string `gorm:"varchar(255);index" json:"name"`
	Current uint64 `json:"height,omitempty"`
	Max     uint64 `json:"max,omitempty"`
	Error   string `gorm:"varchar(255)" json:"error,omitempty"`
}

// PeerSystemInformation for report
type PeerSystemInformation struct {
	gorm.Model `json:"-"`

	Nodes []ChainType `json:"nodes,omitempty"`
	Error string      `gorm:"varchar(255)" json:"error,omitempty"`

	ClientTime int64  `json:"timestamp"`
	ServerTime int64  `json:"servertime"`
	Name       string `gorm:"varchar(255)" json:"name"`

	CPU     []TimesStat        `json:"cpu,omitempty"`
	MemoryV *VirtualMemoryStat `json:"memv,omitempty"`
	MemoryS *SwapMemoryStat    `json:"mems,omitempty"`
	Disk    []UsageStat        `json:"disk,omitempty"`
}

func getPeerSystemTables() []interface{} {
	return []interface{}{
		&TimesStat{},
		&VirtualMemoryStat{},
		&SwapMemoryStat{},
		&UsageStat{},
		&ChainType{},
		&PeerSystemInformation{},
	}
}
