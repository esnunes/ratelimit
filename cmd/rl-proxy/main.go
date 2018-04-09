package main

import (
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/esnunes/ratelimit/pkg/proxy"
)

// NewProxyCmd creates the root proxy command
func NewProxyCmd() *cobra.Command {
	var (
		addr  string
		burst int
		rate  float64
		queue int
	)

	cmd := &cobra.Command{
		Use:   "rl-proxy",
		Short: "Rate Limit Proxy",
		Long:  "Rate Limit Proxy",
	}

	cmd.RunE = func(_ *cobra.Command, args []string) error {
		p := proxy.New(proxy.ServerOptions{
			Burst: burst,
			Rate:  rate,
			Queue: queue,
		})

		log.Printf("rl-proxy: burst [%v], rate [%v], queue [%v], addr [%v]", burst, rate, queue, addr)

		return http.ListenAndServe(addr, p)
	}

	cmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "bind address ip:port")

	cmd.Flags().IntVarP(&burst, "burst", "b", 1, "maximum burst size of requests")
	cmd.Flags().Float64VarP(&rate, "rate", "r", 2, "requests per second")
	cmd.Flags().IntVarP(&queue, "queue", "q", 0, "maximum queued requests")

	return cmd
}

func main() {
	cmd := NewProxyCmd()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
