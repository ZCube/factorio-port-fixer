/*
Copyright © 2023 ZCube <zcubekr@gmail.com>
This file is part of CLI application factorip-port-fixer.
*/
package cmd

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"net"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	pingIp         string
	pingPort       uint16
	pingRemotePort uint16
	pingHostname   string
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "ping",
	Long:  `ping`,
	Run: func(cmd *cobra.Command, args []string) {

		logger, _ := zap.NewProduction()
		defer logger.Sync() // flushes buffer, if any
		sugar := logger.Sugar()

		ip := pingIp
		hostname := pingHostname
		port := pingPort

		addr := net.UDPAddr{IP: net.ParseIP(ip)}

		ips, err := net.LookupIP(hostname)
		if err != nil {
			sugar.Errorf("Could not get IPs: %v", err)
		}
		// TODO: only use one ip ?
		remote := &net.UDPAddr{Port: int(port), IP: ips[0]}

		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			sugar.Fatal(err)
		}

		pktIndex := uint16(rand.Intn(math.MaxUint16 + 1))
		resBuffer := &bytes.Buffer{}

		// ping
		resBuffer.WriteByte(0)
		pktIndexBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(pktIndexBytes, pktIndex)
		resBuffer.Write(pktIndexBytes)

		cc, wrerr := conn.WriteTo(resBuffer.Bytes(), remote)
		if wrerr != nil {
			sugar.Errorf("net.WriteTo() error: %s\n", wrerr)
		} else {
			sugar.Debugw("Wrote to socket",
				"Bytes", cc,
				"Remote", remote)
		}

		b := make([]byte, 2048)

		// pong
		conn.SetReadDeadline(time.Now().Add(time.Second * 5))
		cc, remote, rderr := conn.ReadFromUDP(b)
		if rderr != nil {
			sugar.Fatalf("net.ReadFromUDP() error: %s", rderr)
		} else {
			sugar.Debugw("Read from socket",
				"Bytes", cc,
				"Remote", remote)
		}

	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
	pingCmd.Flags().StringVar(&pingIp, "ip", "0.0.0.0", "ip")
	pingCmd.Flags().StringVar(&pingHostname, "hostname", "localhost", "hostname")
	pingCmd.Flags().Uint16Var(&pingPort, "port", 34197, "port")
}
