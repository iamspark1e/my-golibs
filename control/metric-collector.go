package control

import (
	"context"
	"fmt"
	goNativeNet "net"
	"runtime"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type CPUInfo struct {
	ModelName string
	Cores     int
}

type SystemInfo struct {
	Os             string
	OsVersion      string
	Architecture   string
	Virtualization string
}

type StaticMetrics struct {
	CPU             CPUInfo
	System          SystemInfo
	OuterIPAddr     string
	EnableContainer bool
	MessageType     string
}

type MemInfo struct {
	Total string
	Used  string
	Free  string
}
type SwapInfo struct {
	Total string
	Used  string
	Free  string
}
type Percent struct {
	CPU  string
	Mem  string
	Disk string
	Swap string
}
type Load struct {
	CPU  *load.AvgStat
	Swap SwapInfo
	Mem  MemInfo
}
type Host struct {
	Uptime uint64
}
type InterfaceInfo struct {
	Addrs    []string
	ByteSent uint64
	ByteRecv uint64
}
type DockerContainer struct {
	ID    string
	Name  []string
	State string
}

type DynamicMetrics struct {
	Percent     Percent
	Load        Load
	Host        Host
	Network     map[string]InterfaceInfo
	Container   []DockerContainer
	MessageType string
}

func ListDockerContainers() []DockerContainer {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	err_list := make([]DockerContainer, 0)
	if err != nil {
		return err_list
	}
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return err_list
	}
	list := make([]DockerContainer, len(containers))
	for i, container := range containers {
		list[i].ID = container.ID
		list[i].Name = container.Names
		list[i].State = container.State
	}
	return list
}

// Get preferred outbound ip of this machine
func getOutboundIP() string {
	conn, err := goNativeNet.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*goNativeNet.UDPAddr)
	return localAddr.IP.String()
	// FIXME: another solution, use udp detect cannot break container.
	// req, _ := http.NewRequest("GET", "http://ip.sb/", nil)
	// req.Header.Set("User-Agent", "curl/7.74.0")
	// resp, err := (&http.Client{}).Do(req)
	// if err != nil {
	// 	return ""
	// }
	// defer resp.Body.Close()
	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return ""
	// }
	// bs := string(body)
	// return bs
}

func StaticMetricsData() *StaticMetrics {
	ss := new(StaticMetrics)

	// psutil - cpu
	c, _ := cpu.Info()
	t_cpu := make([]CPUInfo, len(c))
	for i, ci := range c {
		t_cpu[i].ModelName = ci.ModelName
	}
	ss.CPU.ModelName = t_cpu[0].ModelName
	ss.CPU.Cores = len(c)

	// psutil - host
	n, _ := host.Info()
	ss.System.Os = n.Platform
	ss.System.OsVersion = n.PlatformVersion
	ss.System.Architecture = runtime.GOARCH
	ss.System.Virtualization = n.VirtualizationSystem

	ss.OuterIPAddr = strings.ReplaceAll(getOutboundIP(), "\n", "")

	ss.MessageType = "info"
	ss.EnableContainer = true

	return ss
}

func DynamicMetricsData() *DynamicMetrics {
	cc, _ := cpu.Percent(time.Second, false)
	v, _ := mem.VirtualMemory()
	vs, _ := mem.SwapMemory()
	d, _ := disk.Usage("/")
	n, _ := host.Info()
	nv, _ := net.IOCounters(true)

	ss := new(DynamicMetrics)
	ss.Percent.CPU = fmt.Sprintf("%.3g", cc[0])
	ss.Percent.Mem = fmt.Sprintf("%.3g", v.UsedPercent)
	ss.Percent.Disk = fmt.Sprintf("%.3g", d.UsedPercent)
	ss.Percent.Swap = fmt.Sprintf("%.3g", vs.UsedPercent)

	t_cpu_load, _ := load.Avg()
	ss.Load.CPU = t_cpu_load
	ss.Load.Swap.Total = fmt.Sprintf("%d", vs.Total/1024/1024)
	ss.Load.Swap.Used = fmt.Sprintf("%d", vs.Used/1024/1024)
	ss.Load.Swap.Free = fmt.Sprintf("%d", vs.Free/1024/1024)
	ss.Load.Mem.Total = fmt.Sprintf("%d", v.Total/1024/1024)
	ss.Load.Mem.Used = fmt.Sprintf("%d", v.Used/1024/1024)
	ss.Load.Mem.Free = fmt.Sprintf("%d", v.Free/1024/1024)

	ss.Host.Uptime = n.Uptime

	ss.Network = make(map[string]InterfaceInfo)
	for _, v := range nv {
		if v.Name != "lo" {
			var ii InterfaceInfo
			ii.ByteSent = v.BytesSent
			ii.ByteRecv = v.BytesRecv
			ss.Network[v.Name] = ii
		}
	}
	ss.Container = ListDockerContainers()
	ss.MessageType = "state"

	return ss
}
