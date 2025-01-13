package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/quicvarint"
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

var totalSize uint64

func handleConn(ctx context.Context, conn quic.Connection) {
	for {
		stream, err := conn.AcceptUniStream(ctx)
		if err != nil {
			log.Print(err)
			return
		}
		go func() {
			written, err := io.Copy(io.Discard, stream)
			if err != nil {
				log.Print(err)
				return
			}
			atomic.AddUint64(&totalSize, uint64(written))
		}()
	}
}

func runServer(ctx context.Context, listenAddr string) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quicConfig := &quic.Config{
		MaxIncomingUniStreams:      1 << 60,
		InitialStreamReceiveWindow: 1232,
		MaxStreamReceiveWindow:     quicvarint.Max,
		MaxIdleTimeout:             10 * time.Second,
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{genSolanaCert()},
		NextProtos:         []string{"solana-tpu"},
		InsecureSkipVerify: true,
	}

	go func() {
		timer := time.NewTicker(time.Second)
		defer timer.Stop()
		var lastRcvd uint64
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				curRcvd := atomic.LoadUint64(&totalSize)
				diff := curRcvd - lastRcvd
				fmt.Printf("%.3f Mbps\n", (float64(diff)*8)/1e6)
				lastRcvd = curRcvd
			}
		}
	}()

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
