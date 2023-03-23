package cmd

import (
	"net"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/leosocy/proksi/pkg/middleman"
	"github.com/leosocy/proksi/pkg/sched"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Start proksi server, act as middleman or datasource",
	Long:    "",
	Aliases: []string{"serve", "srv"},
}

type serverMitmCmd struct {
	bind string
	port int

	cmd *cobra.Command
}

func (cc *serverMitmCmd) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mitm",
		Short: "Start proksi middleman server",
		Long: `The client can configure the listening port of mitm as an http(s) proxy server,
and proksi will forward the request to the best proxy server`,
		RunE:    cc.runE,
		Aliases: []string{"middleman"},
	}

	cmd.Flags().StringVarP(&cc.bind, "bind", "b", "0.0.0.0", "interface to which the server will bind")
	cmd.Flags().IntVarP(&cc.port, "port", "p", 8081, "port on which the server will listen")

	cc.cmd = cmd
	return cc.cmd
}

func (cc *serverMitmCmd) runE(cmd *cobra.Command, args []string) error {
	scheduler := sched.NewScheduler()
	go scheduler.Start()

	middlemanServer := middleman.NewServer(scheduler.GetBackend())
	addr := net.JoinHostPort(cc.bind, strconv.Itoa(cc.port))
	http.ListenAndServe(addr, middlemanServer)
	return nil
}
