package machine

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
	"github.com/whoisfisher/mykubespray/pkg/utils"
	"sync"
)

type Machine struct {
	ID         int
	IsWorking  bool
	FreeMemory string
	FreeCPU    string
	Process    string
	Host       entity.Host
}

type MachineManager struct {
	observers     sync.Map
	machines      map[int]Machine
	osClients     map[int]*utils.OSClient
	updateChannel chan Machine
}

type AdvancedObserver struct {
	ID                 int
	SubscribedMachines []int
}
