package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/logging"
	"github.com/quic-go/quic-go/qlog"
	"github.com/spf13/cobra"
)

var connCheckCmd = cobra.Command{
	Use:   "conn-check",
	Short: "Connect and dump qlog trace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runConnCheck(context.Background(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(&connCheckCmd)
}

func runConnCheck(ctx context.Context, dst string) {
	quicConfig := &quic.Config{
		Tracer: func(ctx context.Context, p logging.Perspective, odcid quic.ConnectionID) *logging.ConnectionTracer {
			return qlog.NewConnectionTracer(os.Stdout, p, odcid)
		},
	}

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
}
