package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"

	"github.com/quic-go/quic-go"
	"github.com/spf13/cobra"
)

var serverCmd = cobra.Command{
	Use:   "server <endpoint:port>",
	Short: "Run QUIC server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runServer(context.Background(), args[0])
	},
}

func init() {
	rootCmd.AddCommand(&serverCmd)
}

func handleConn(ctx context.Context, conn quic.Connection) {
	for {
		stream, err := conn.AcceptUniStream(ctx)
		if err != nil {
			log.Print(err)
			return
		}
		io.Copy(io.Discard, stream)
	}
}

func runServer(ctx context.Context, listenAddr string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quicConfig := &quic.Config{}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{genSolanaCert()},
		NextProtos:         []string{"solana-tpu"},
		InsecureSkipVerify: true,
	}

	listener, err := quic.ListenAddr(listenAddr, tlsConfig, quicConfig)
	if err != nil {
		log.Print(err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on", listenAddr)
	for {
		conn, err := listener.Accept(ctx)
		if err != nil {
			log.Print(err)
			return
		}
		go handleConn(ctx, conn)
	}
}
