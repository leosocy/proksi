package main

import (
	"github.com/Leosocy/IntelliProxy/pkg/sched"
)

func main() {
	scheduler := sched.NewScheduler()
	// scheduler.RateLimit(&sched.LimitRule{Delay: 5 * time.Second, Parallelism: 10}) // ≈ 60/Delay*Parall=120times/min
	scheduler.Start()
}
