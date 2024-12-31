package machine

type Observer interface {
	Update(machine Machine)
}

type Subject interface {
	RegisterObserver(observer Observer, machineID int)
	RemoveObserver(observer Observer, machineID int)
	NotifyObservers(machineID int, isWorking bool, freeMemory int, freeCPU float64)
}
