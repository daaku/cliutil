package cliutil

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// MustEnv returns the environment variable for the given key, and panics if it
// isn't found.
func MustEnv(k string) string {
	v, found := os.LookupEnv(k)
	if !found {
		panic(fmt.Sprintf("%q must be set", k))
	}
	return v
}

// Run the function until completion, or cancel if SIGTERM or SIGINT is
// recieved.
func Run(run func(context.Context) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	res := make(chan error, 1)
	go func() { res <- run(ctx) }()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	select {
	case err := <-res:
		signal.Stop(sigs)
		return err
	case <-sigs:
		signal.Stop(sigs)
		cancel()
		select {
		case err := <-res:
			signal.Stop(sigs)
			return err
		case <-time.After(10 * time.Second):
			return errors.New("cliutil: timeout in cancelling run")
		}
	}
}
