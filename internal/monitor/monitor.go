package monitor

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type ResourceUsage struct {
	CPU       float64   `json:"cpu"`
	MemoryMB  uint64    `json:"memory_mb"`
	GPU       *GPUUsage `json:"gpu,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type GPUUsage struct {
	Index     int     `json:"index"`
	Name      string  `json:"name"`
	Utilization float64 `json:"utilization"`
	MemoryUsedMB uint64 `json:"memory_used_mb"`
	MemoryTotalMB uint64 `json:"memory_total_mb"`
	Temperature int     `json:"temperature"`
	PowerWatts float64  `json:"power_watts"`
}

type Monitor struct {
	interval time.Duration
	stopCh   chan struct{}
}

func NewMonitor(interval time.Duration) *Monitor {
	return &Monitor{
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (m *Monitor) GetUsage() (*ResourceUsage, error) {
	usage := &ResourceUsage{
		Timestamp: time.Now(),
	}

	switch runtime.GOOS {
	case "windows":
		if err := m.collectWindows(usage); err != nil {
			return nil, err
		}
	default:
		if err := m.collectLinux(usage); err != nil {
			return nil, err
		}
	}

	return usage, nil
}

func (m *Monitor) collectWindows(usage *ResourceUsage) error {
	cpu, err := m.getWindowsCPU()
	if err == nil {
		usage.CPU = cpu
	}

	memoryMB, err := m.getWindowsMemory()
	if err == nil {
		usage.MemoryMB = memoryMB
	}

	gpu, err := m.getWindowsGPU()
	if err == nil {
		usage.GPU = gpu
	}

	return nil
}

func (m *Monitor) collectLinux(usage *ResourceUsage) error {
	cpu, err := m.getLinuxCPU()
	if err == nil {
		usage.CPU = cpu
	}

	memoryMB, err := m.getLinuxMemory()
	if err == nil {
		usage.MemoryMB = memoryMB
	}

	gpu, err := m.getLinuxGPU()
	if err == nil {
		usage.GPU = gpu
	}

	return nil
}

func (m *Monitor) getWindowsCPU() (float64, error) {
	cmd := exec.Command("powershell", "-Command",
		"(Get-CimInstance Win32_Processor).LoadPercentage")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (m *Monitor) getWindowsMemory() (uint64, error) {
	cmd := exec.Command("powershell", "-Command",
		"$os = Get-CimInstance Win32_OperatingSystem; [math]::Round(($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / 1KB)")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return uint64(val), nil
}

func (m *Monitor) getWindowsGPU() (*GPUUsage, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no GPU data")
	}

	return parseGPULine(lines[0])
}

func (m *Monitor) getLinuxCPU() (float64, error) {
	cmd := exec.Command("top", "-bn1", "-w", "512")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Cpu(s)") {
			fields := strings.Fields(line)
			for i, f := range fields {
				if f == "id," && i+1 < len(fields) {
					idle, err := strconv.ParseFloat(strings.Trim(fields[i+1], "%"), 64)
					if err == nil {
						return 100.0 - idle, nil
					}
				}
			}
		}
	}
	return 0, fmt.Errorf("could not parse CPU usage")
}

func (m *Monitor) getLinuxMemory() (uint64, error) {
	cmd := exec.Command("free", "-m")
	out, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				used, err := strconv.ParseUint(fields[2], 10, 64)
				if err == nil {
					return used, nil
				}
			}
		}
	}
	return 0, fmt.Errorf("could not parse memory usage")
}

func (m *Monitor) getLinuxGPU() (*GPUUsage, error) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,utilization.gpu,memory.used,memory.total,temperature.gpu,power.draw",
		"--format=csv,noheader,nounits")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no GPU data")
	}

	return parseGPULine(lines[0])
}

func parseGPULine(line string) (*GPUUsage, error) {
	fields := strings.Split(line, ",")
	if len(fields) < 7 {
		return nil, fmt.Errorf("invalid GPU line: %s", line)
	}

	index, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
	name := strings.TrimSpace(fields[1])
	util, _ := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
	memUsed, _ := strconv.ParseUint(strings.TrimSpace(fields[3]), 10, 64)
	memTotal, _ := strconv.ParseUint(strings.TrimSpace(fields[4]), 10, 64)
	temp, _ := strconv.Atoi(strings.TrimSpace(fields[5]))
	power, _ := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64)

	return &GPUUsage{
		Index:         index,
		Name:          name,
		Utilization:   util,
		MemoryUsedMB:  uint64(memUsed),
		MemoryTotalMB: uint64(memTotal),
		Temperature:   temp,
		PowerWatts:    power,
	}, nil
}

func (m *Monitor) StartStreaming(callback func(*ResourceUsage)) {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				usage, err := m.GetUsage()
				if err == nil {
					callback(usage)
				}
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}
