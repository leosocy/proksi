package main

import (
	"github.com/leosocy/proksi/pkg/middleman"
	"net/http"

	"github.com/leosocy/proksi/pkg/sched"
)

func main() {
	scheduler := sched.NewScheduler()
	go scheduler.Start()

	middlemanServer := middleman.NewServer(scheduler.GetBackend())
	http.ListenAndServe("0.0.0.0:8081", middlemanServer)
}
