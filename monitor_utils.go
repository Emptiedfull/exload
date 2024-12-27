package main

import (
	"fmt"
	"os"

	"github.com/shirou/gopsutil/process"
)

func memInfo(p *pen) int {

	mem := 0

	for _, srv := range p.servers {

		pro, err := process.NewProcess(int32(srv.cmd.Process.Pid))
		if err != nil {
			fmt.Println("Error reading mem")
			continue
		}
		memIn, err := pro.MemoryInfo()
		mem += int(memIn.RSS)
		if err != nil {
			fmt.Println("Error reading mem")
			continue
		}

		children, err := pro.Children()
		if err != nil {
			fmt.Println("Error reading chilren")
			continue
		}
		for _, child := range children {
			childMem, err := child.MemoryInfo()
			if err != nil {
				continue
			}
			mem += int(childMem.RSS)
		}

	}

	return mem
}

func totalMem(u map[string]*pen) int {
	mem := 0
	for _, pen := range u {
		mem += memInfo(pen)

	}
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		fmt.Println("error reading parent mem")
		return mem
	}
	m, err := p.MemoryInfo()
	if err != nil {
		fmt.Println("Error reading parent mem")
		return mem
	}

	mem += int(m.RSS)

	memMB := mem / 1024 / 1024

	return memMB
}

func getActiveServersByPen(p *pen) int {
	return len(p.servers)
}

func getActiveServers(u map[string]*pen) int {
	srvs := 0
	for _, pen := range u {
		srvs += getActiveServersByPen(pen)
	}
	return srvs
}

func getTotalPorts(m *manager) (int, int) {
	p := len(m.free_ports)
	s := getActiveServers(m.UrlMap)

	return p + s, s
}

func getRpsByServer(s *server) int {
	res := 0
	s.reqMu.Lock()
	res += s.req
	s.req = 0
	s.reqMu.Unlock()
	return res
}

func getRpsByPen(p *pen) int {
	res := 0
	for _, srv := range p.servers {
		res += getRpsByServer(srv)

	}
	return res
}

func TotalRps(u map[string]*pen) int {
	res := 0
	for _, pen := range u {
		res += getRpsByPen(pen)
	}
	return res
}
