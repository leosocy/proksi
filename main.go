package main

import (
	"github.com/Leosocy/IntelliProxy/pkg/sched"
	"github.com/Leosocy/IntelliProxy/service/middleman"
	"net/http"
)

func main() {
	scheduler := sched.NewScheduler()

	middlemanServer := middleman.NewServer(scheduler.GetStorage())
	go http.ListenAndServe("0.0.0.0:8081", middlemanServer)

	scheduler.Start()
}
