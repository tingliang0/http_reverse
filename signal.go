package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init_signals() {
	g_sigs := make(chan os.Signal, 1)
	signal.Notify(g_sigs, syscall.SIGUSR1)
	go func() {
		for {
			sig := <-g_sigs
			log.Printf("receive sig: %+v\n", sig)
			load_cfg()
		}
	}()
}
