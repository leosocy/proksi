package main

import (
	"github.com/Leosocy/IntelliProxy/service/middleman"
)

func main() {
	//scheduler := sched.NewScheduler()
	//scheduler.Start()
	middleman.ListenAndServe()
}
