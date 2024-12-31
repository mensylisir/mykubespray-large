package machine

import (
	"github.com/whoisfisher/mykubespray/pkg/entity"
)

func main() {
	// 创建五台机器的初始状态
	machines := []Machine{
		{ID: 1,
			Process: "kk",
			Host: entity.Host{
				Name:     "aaa",
				Address:  "192.168.227.160",
				Port:     22,
				User:     "wangzhendong",
				Password: "Def@u1tpwd",
			}}, {ID: 2,
			Process: "kk",
			Host: entity.Host{
				Name:     "bbb",
				Address:  "192.168.227.161",
				Port:     22,
				User:     "wangzhendong",
				Password: "Def@u1tpwd",
			}}, {ID: 3,
			Process: "kk",
			Host: entity.Host{
				Name:     "ccc",
				Address:  "192.168.227.162",
				Port:     22,
				User:     "wangzhendong",
				Password: "Def@u1tpwd",
			}},
	}

	manager := NewMachineManager(machines)

	// 创建观察者
	observer1 := &AdvancedObserver{ID: 1}
	observer2 := &AdvancedObserver{ID: 2}

	// 观察者1订阅机器1和机器2的状态变化通知
	observer1.SubscribeMachine(1)
	observer1.SubscribeMachine(2)

	// 观察者2订阅机器3的状态变化通知
	observer2.SubscribeMachine(3)

	// 注册观察者
	manager.RegisterObserver(observer1, 1)
	manager.RegisterObserver(observer1, 2)
	manager.RegisterObserver(observer2, 3)

	select {}
}
