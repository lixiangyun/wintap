package main

import (
	"fmt"
	"github.com/lixiangyun/wintap"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"os/signal"
)

func WinTunRecvFunc(tun *wintap.Tun, buff []byte)  {
	ipTye := wintap.IPHeaderType(buff[0])
	if ipTye == wintap.IPv4 {
		ip4hdr := wintap.IP4HeaderDecoder(buff[:wintap.MAX_IPHEADER])
		fmt.Printf("IPv4 Packet: %s\n", ip4hdr.String())
		return
	}

	if ipTye == wintap.IPv6 {
		/*ip6hdr := wintap.IP6HeaderDecoder(buff[:wintap.MAX_IP6HEADER])
		fmt.Printf("IPv6 Packet: %s\n", ip6hdr.String())*/
		return
	}

	fmt.Printf("unkown packet %d:%v\n", len(buff), buff)
}

func main() {
	tun, err := wintap.OpenTun([]byte{192, 172, 3, 3}, []byte{192, 172, 3, 0}, []byte{255, 255, 255, 0})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = tun.SetDHCPMasq(
		net.IP([]byte{192, 172, 3, 3}),
		net.IP([]byte{255, 255, 255, 0}),
		net.IP([]byte{0, 0, 0, 0}),
		365*24*3600 )
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("windows tun init success\n")

	err = tun.Connect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("mtu: %d\n", tun.GetMTU(false))
	time.Sleep(2 * time.Second)

	err = tun.SetReadHandler(WinTunRecvFunc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wp := sync.WaitGroup{}
	wp.Add(1)
	go func () {
		fmt.Println("windows tun run....")
		tun.Listen(10)
		wp.Done()
	} ()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)

	signals := <- signalChan
	fmt.Printf("windows tun recv singal %s\n", signals.String())

	tun.SignalStop()
	wp.Wait()
}
