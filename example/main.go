package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"go.acpr.dev/touchportal-golang-sdk/pkg/plugin"
)

type settings struct {
	Host string `json:"Host"`
	Port int    `json:"Port"`
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	defer func() {
		signal.Stop(c)
		cnl()
	}()
	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Quitting...\n", sig)
			cnl()
		case <-ctx.Done():
		}
	}()

	s := &settings{}
	p := plugin.NewPlugin(ctx, "gsdk")
	p.Settings(s)
	p.Register()

	// if you want an easy way to wait around for the plugin to exit plugin.Done() offers
	// an unbuffered channel you can watch.
	<-p.Done()
}
