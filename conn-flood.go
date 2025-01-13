package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var connFloodCmd = cobra.Command{
	Use:   "conn-flood <host:port>",
	Short: "Flood endpoint with uni streams",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runConnFlood(context.Background(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(&connFloodCmd)
}

func runConnFlood(ctx context.Context, dst string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var totalSent uint64

	go func() {
		timer := time.NewTicker(time.Second)
		defer timer.Stop()
		var lastSent uint64
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				curSent := atomic.LoadUint64(&totalSent)
				diff := curSent - lastSent
				fmt.Printf("%g\n", float64(diff))
				lastSent = curSent
			}
		}
	}()

	quicConfig := &quic.Config{}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{genSolanaCert()},
		NextProtos:         []string{"solana-tpu"},
		InsecureSkipVerify: true,
	}

	conn, err := quic.DialAddr(ctx, dst, tlsConfig, quicConfig)
	if err != nil {
		log.Print(err)
		return
	}
	defer conn.CloseWithError(0, "")

	sem := semaphore.NewWeighted(65536)

	var payload [1232]byte
	for {
		stream, err := conn.OpenUniStream()
		if err != nil {
			log.Printf("%#v", err)
			return
		}
		sem.Acquire(ctx, 1)
		go func(stream quic.SendStream) {
			defer sem.Release(1)
			n, err := stream.Write(payload[:])
			if err != nil {
				log.Printf("%#v", err)
				return
			}
			stream.Close()

			atomic.AddUint64(&totalSent, uint64(n))
		}(stream)
	}
}
