/*
Copyright © 2023 ZCube <zcubekr@gmail.com>
This file is part of CLI application factorip-port-fixer.
*/
package cmd

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

var (
	remoteIp         string
	remotePort       uint16
	remoteRemotePort uint16
	remoteHostname   string
)

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "remote mode",
	Long:  `remote mode`,
	Run: func(cmd *cobra.Command, args []string) {

		logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		ip := remoteIp
		port := remotePort
		remotePort := remoteRemotePort

		addr := net.UDPAddr{Port: int(port), IP: net.ParseIP(ip)}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			sugar.Fatal(err)
		}

		b := make([]byte, 2048)

		mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGKILL, os.Interrupt)
		defer stop()

		g, gCtx := errgroup.WithContext(mainCtx)

		g.Go(func() error {
			for {
				sugar.Infow("Accepting a new packet",
					"IP", ip,
					"Port", port)

				cc, remote, rderr := conn.ReadFromUDP(b)
				if rderr != nil {
					sugar.Errorf("net.ReadFromUDP() error: %s", rderr)
					return rderr
				} else {
					sugar.Infow("Read from socket",
						"Bytes", cc,
						"Remote", remote)
				}

				req := b[:cc]
				if len(req) < 3 {
					// ignore keepalive packet
					continue
				}
				pktIndex := binary.LittleEndian.Uint16(req[1:])

				resBuffer := &bytes.Buffer{}
				// pong
				resBuffer.WriteByte(9)
				pktIndexBytes := make([]byte, 2)
				binary.LittleEndian.PutUint16(pktIndexBytes, pktIndex)
				resBuffer.Write(pktIndexBytes)

				address := fmt.Sprintf("%v:%v", remote.IP, remotePort)

				addressLength := uint32(len(address))
				addressLengthBytes := make([]byte, 4)
				binary.LittleEndian.PutUint32(addressLengthBytes, addressLength)
				resBuffer.Write(addressLengthBytes)

				resBuffer.Write([]byte(address))

				cc, wrerr := conn.WriteTo(resBuffer.Bytes(), remote)
				if wrerr != nil {
					sugar.Errorf("net.WriteTo() error: %s\n", wrerr)
				} else {
					sugar.Infow("Wrote to socket",
						"Bytes", cc,
						"Remote", remote)
				}
			}
		})

		g.Go(func() error {
			<-gCtx.Done()
			_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			conn.Close()
			sugar.Infow("graceful shutting down")
			return nil
		})

		if err := g.Wait(); err != nil {
			sugar.Errorf("exit reason: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(remoteCmd)
	remoteCmd.Flags().StringVar(&remoteIp, "ip", "0.0.0.0", "ip")
	remoteCmd.Flags().Uint16Var(&remotePort, "port", 34197, "port")
	remoteCmd.Flags().Uint16Var(&remoteRemotePort, "remotePort", 34197, "remote port")
}
