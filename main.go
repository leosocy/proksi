package main

import (
	"net/http"

	"github.com/Leosocy/IntelliProxy/pkg/sched"
	"github.com/Leosocy/IntelliProxy/service/middleman"
)

func main() {
	scheduler := sched.NewScheduler()
	go scheduler.Start()

	middlemanServer := middleman.NewServer(scheduler.GetBackend())
	http.ListenAndServe("0.0.0.0:8081", middlemanServer)
}
