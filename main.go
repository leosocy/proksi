package main

import (
	"github.com/Leosocy/IntelliProxy/pkg/sched"
)

func main() {
	scheduler := sched.NewScheduler()
	scheduler.Start()
}
