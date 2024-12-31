package machine

import "fmt"

func (ao *AdvancedObserver) Update(machine Machine) {
	if ao.shouldNotify(machine.ID) {
		fmt.Printf("Observer %d 收到机器状态变化事件：机器ID %d，工作状态 %t，空闲内存 %s，空闲CPU %s\n", ao.ID, machine.ID, machine.IsWorking, machine.FreeMemory, machine.FreeCPU)
	}
}

func (ao *AdvancedObserver) SubscribeMachine(machineID int) {
	for _, id := range ao.SubscribedMachines {
		if id == machineID {
			return // 已经订阅过该机器
		}
	}
	ao.SubscribedMachines = append(ao.SubscribedMachines, machineID)
}

func (ao *AdvancedObserver) UnsubscribeMachine(machineID int) {
	for i, id := range ao.SubscribedMachines {
		if id == machineID {
			ao.SubscribedMachines = append(ao.SubscribedMachines[:i], ao.SubscribedMachines[i+1:]...)
			return
		}
	}
}

func (ao *AdvancedObserver) shouldNotify(machineID int) bool {
	if len(ao.SubscribedMachines) == 0 {
		return true // 未订阅任何机器，默认接收所有通知
	}
	for _, id := range ao.SubscribedMachines {
		if id == machineID {
			return true
		}
	}
	return false
}
