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
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

		signal.Ignore(
			syscall.SIGHUP,
		)
		mainCtx, stop := signal.NotifyContext(context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
			syscall.SIGKILL,
		)
		defer stop()

		g, gCtx := errgroup.WithContext(mainCtx)

		e := echo.New()

		e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
			LogURI:    true,
			LogStatus: true,
			LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
				if v.Error != nil {
					sugar.Errorw("request",
						"URI", v.URI,
						"status", v.Status,
						"RemoteIP", c.RealIP(),
						"Error", v.Error,
					)
				} else {
					sugar.Infow("request",
						"URI", v.URI,
						"status", v.Status,
						"RemoteIP", c.RealIP(),
					)
				}

				return nil
			},
		}))
		e.Use(middleware.Recover())

		e.GET("/health", func(c echo.Context) error {
			{
				remoteClientIp := c.RealIP()

				addr := net.UDPAddr{IP: net.ParseIP(ip)}

				remote := &net.UDPAddr{Port: int(port), IP: net.ParseIP("127.0.0.1")}

				conn, err := net.ListenUDP("udp", &addr)
				if err != nil {
					sugar.Error(err)
					return c.String(http.StatusInternalServerError, err.Error())
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
						"Remote", remote,
						"RemoteClientIP", remoteClientIp,
					)
				}

				b := make([]byte, 2048)

				// pong
				conn.SetReadDeadline(time.Now().Add(time.Second * 5))
				cc, remote, rderr := conn.ReadFromUDP(b)
				if rderr != nil {
					sugar.Errorw(fmt.Sprintf("net.ReadFromUDP() error: %s", rderr),
						"Remote", remote,
					)
					return c.String(http.StatusBadGateway, rderr.Error())
				} else {
					sugar.Debugw("Read from socket",
						"Bytes", cc,
						"Remote", remote)
				}
			}
			return c.String(http.StatusOK, "OK")
		})

		// factorio server health check : from factorio server to this server
		e.GET("/health_for_factorio", func(c echo.Context) error {
			{
				remoteClientIp := c.RealIP()

				addr := net.UDPAddr{IP: net.ParseIP(ip)}

				remote := &net.UDPAddr{Port: int(remotePort), IP: net.ParseIP(remoteClientIp)}

				conn, err := net.ListenUDP("udp", &addr)
				if err != nil {
					sugar.Error(err)
					return c.String(http.StatusInternalServerError, err.Error())
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
						"Remote", remote,
						"RemoteClientIP", remoteClientIp,
					)
				}

				b := make([]byte, 2048)

				// pong
				conn.SetReadDeadline(time.Now().Add(time.Second * 5))
				cc, remote, rderr := conn.ReadFromUDP(b)
				if rderr != nil {
					sugar.Errorw(fmt.Sprintf("net.ReadFromUDP() error: %s", rderr),
						"Remote", remote,
					)
					return c.String(http.StatusBadGateway, rderr.Error())
				} else {
					sugar.Debugw("Read from socket",
						"Bytes", cc,
						"Remote", remote)
				}
			}
			return c.String(http.StatusOK, "OK")
		})

		g.Go(func() error {
			sugar.Info("Healthcheck server started")
			e.HideBanner = true
			if err := e.Start(fmt.Sprintf(":%v", port)); err != nil && err != http.ErrServerClosed {
				sugar.Errorf("shutting down the server : %s", err)
				return err
			}
			return nil
		})

		g.Go(func() error {
			for {
				sugar.Debugw("Accepting a new packet",
					"IP", ip,
					"Port", port)

				cc, remote, rderr := conn.ReadFromUDP(b)
				if rderr != nil {
					sugar.Errorw(fmt.Sprintf("net.ReadFromUDP() error: %s", rderr),
						"Remote", remote,
					)
					return rderr
				} else {
					sugar.Debugw("Read from socket",
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
					sugar.Debugw("Wrote to socket",
						"Bytes", cc,
						"Remote", remote)
				}
			}
		})

		g.Go(func() error {
			<-gCtx.Done()
			sugar.Errorw("shutting down the server",
				"reason", gCtx.Err())

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			conn.Close()
			if err := e.Shutdown(ctx); err != nil {
				sugar.Error(err)
				return err
			}
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
