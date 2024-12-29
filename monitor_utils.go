package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/a-h/templ"
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
	p := 7
	s := getActiveServers(m.UrlMap)

	return p + s, s
}

func getRpsByServer(s *server) int {
	res := 0
	s.rpsMu.RLock()
	res += s.rps
	s.rpsMu.RUnlock()
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

func getConns(u map[string]*pen) int {
	res := 0
	for _, pen := range u {
		res += pen.con
	}
	return res
}

type penFormatted struct {
	name   string
	max    string
	active string
	cmd    string
}

func getPenFormatted(m *manager) []penFormatted {
	res := make([]penFormatted, 0)
	for pre, pen := range m.UrlMap {
		cmd := pen.command.com + " " + strings.Join(pen.command.args, " ")
		active := strconv.Itoa(len(pen.servers))
		max := strconv.Itoa(int(pen.max_servers))
		res = append(res, penFormatted{pre, max, active, cmd})
	}
	return res
}

func createPayload(pS *State, nS State) []templ.Component {
	payload := make([]templ.Component, 0)
	if pS.rps != nS.rps {
		pS.rps = nS.rps
		payload = append(payload, update_status_item("rps", nS.rps))
	}
	if pS.mem != nS.mem {
		pS.mem = nS.mem
		payload = append(payload, update_status_item("mem", nS.mem))
	}
	if pS.srvs != nS.srvs {
		pS.srvs = nS.srvs
		payload = append(payload, update_status_item("srvs", nS.srvs))
	}
	if pS.cons != nS.cons {
		pS.cons = nS.cons
		payload = append(payload, update_status_item("cons", nS.cons))
	}
	return payload
}
