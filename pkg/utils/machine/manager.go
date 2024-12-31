package machine

import (
	"github.com/whoisfisher/mykubespray/pkg/logger"
	"github.com/whoisfisher/mykubespray/pkg/utils"
	"sync"
	"time"
)

func NewMachineManager(machines []Machine) *MachineManager {
	mm := &MachineManager{
		machines:      make(map[int]Machine),
		osClients:     make(map[int]*utils.OSClient),
		updateChannel: make(chan Machine),
	}

	proccessName := ""

	for _, machine := range machines {
		mm.machines[machine.ID] = machine
		osClient := utils.NewClient(machine.Host)
		mm.osClients[machine.ID] = osClient
		proccessName = machine.Process
	}

	go mm.handleUpdates()

	go mm.periodicallyUpdateMachineStatus(proccessName)

	return mm
}

func (mm *MachineManager) RegisterObserver(observer Observer, machineID int) {
	mm.observers.LoadOrStore(machineID, &sync.Map{})
	obs, _ := mm.observers.Load(machineID)
	obs.(*sync.Map).Store(observer, struct{}{})
}

func (mm *MachineManager) RemoveObserver(observer Observer, machineID int) {
	obs, ok := mm.observers.Load(machineID)
	if !ok {
		return
	}
	obs.(*sync.Map).Delete(observer)
}

func (mm *MachineManager) NotifyObservers(machine Machine) {
	obs, ok := mm.observers.Load(machine.ID)
	if !ok {
		return
	}

	obs.(*sync.Map).Range(func(key, value interface{}) bool {
		observer := key.(Observer)
		observer.Update(machine)
		return true
	})
}

func (mm *MachineManager) periodicallyUpdateMachineStatus(processName string) {
	ticker := time.NewTicker(10 * time.Second) // 每隔 10 秒更新一次状态
	defer ticker.Stop()

	for range ticker.C {
		for machineID := range mm.machines {
			mm.UpdateMachineStatus(machineID, processName)
		}
	}
}

func (mm *MachineManager) UpdateMachineStatus(id int, processName string) {
	machine, ok := mm.machines[id]
	if !ok {
		logger.GetLogger().Errorf("Machine with machine ID %d not found", id)
		return
	}

	client, ok := mm.osClients[id]
	if !ok {
		logger.GetLogger().Errorf("SSH client not found for machine ID %d", id)
		return
	}

	go func() {
		freeMemory := client.GetAvailableMemory()
		freeCPU := client.GetAvailableCPU()
		status := client.IsProcessExist(processName)
		mm.updateChannel <- Machine{
			ID:         machine.ID,
			IsWorking:  status,
			FreeMemory: freeMemory,
			FreeCPU:    freeCPU,
		}
	}()
}

func (mm *MachineManager) handleUpdates() {
	for update := range mm.updateChannel {
		mm.NotifyObservers(update)
	}
}
