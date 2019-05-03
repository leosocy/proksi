package main

import "github.com/Leosocy/gipp/pkg/sched"

func main() {
	scheduler := sched.NewScheduler()
	scheduler.Start()
}
