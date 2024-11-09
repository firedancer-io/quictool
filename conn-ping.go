package main

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/spf13/cobra"
)

var connPingCmd = cobra.Command{
	Use:   "conn-ping",
	Short: "Ping connection creation",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runConnPing(context.Background(), args[0])
	},
}

var connPingFlags struct {
	interval time.Duration
}

func init() {
	rootCmd.AddCommand(&connPingCmd)

	flags := connPingCmd.Flags()
	flags.DurationVarP(&connPingFlags.interval, "interval", "i", 1*time.Second, "Ping interval")
}

func runConnPing(ctx context.Context, dst string) {
	ticker := time.NewTicker(connPingFlags.interval)
	defer ticker.Stop()
	for range ticker.C {
		go pingOnce(ctx, dst)
	}
}

func pingOnce(ctx context.Context, dst string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	now := time.Now()
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{genSolanaCert()},
		NextProtos:         []string{"solana-tpu"},
		InsecureSkipVerify: true,
	}
	conn, err := quic.DialAddr(ctx, dst, tlsConfig, nil)
	if err != nil {
		log.Print(err)
		return
	}
	defer conn.CloseWithError(0, "")
	log.Print(time.Since(now))
}
