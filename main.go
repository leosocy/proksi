package main

import (
	"net/http"

	"github.com/Leosocy/IntelliProxy/pkg/sched"
	"github.com/Leosocy/IntelliProxy/service/middleman"
)

func main() {
	scheduler := sched.NewScheduler()

	middlemanServer := middleman.NewServer(scheduler.GetStorage())
	http.ListenAndServe("0.0.0.0:8081", middlemanServer)

	//scheduler.Start()
}
