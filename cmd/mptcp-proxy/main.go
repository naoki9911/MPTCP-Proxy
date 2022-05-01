package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

const IPPROTO_MPTCP = 262

var (
	remoteAddr net.IP
	remotePort int
	localPort  int
	mode       string
)

func main() {
	if log.GetLevel() == log.DebugLevel {
		log.SetReportCaller(true)
	}

	var err error

	flag.StringVarP(&mode, "mode", "m", "", "specify mode (server or client)")
	flag.IntVarP(&localPort, "port", "p", 0, "local bind port")
	var rAddr *string = flag.StringP("remote", "r", "", "remote address (ex. 127.0.0.1:8080)")
	flag.Parse()

	if localPort == 0 || *rAddr == "" || mode == "" {
		flag.Usage()
		return
	}

	if mode != "server" && mode != "client" {
		flag.Usage()
		return
	}

	addrs := strings.Split(*rAddr, ":")
	if len(addrs) != 2 {
		flag.Usage()
		return
	}

	remoteAddr = net.ParseIP(addrs[0])
	if remoteAddr == nil {
		resolvedAddr, err := net.ResolveIPAddr("ip4", addrs[0])
		if err != nil {
			log.Error(err)
			flag.Usage()
			return
		}
		remoteAddr = resolvedAddr.IP.To4()
	}

	remotePort, err = strconv.Atoi(addrs[1])
	if err != nil {
		flag.Usage()
		return
	}

	log.Infof("starting proxy...")
	if mode == "client" {
		doProxy(syscall.IPPROTO_IP, IPPROTO_MPTCP)
	} else if mode == "server" {
		doProxy(IPPROTO_MPTCP, syscall.IPPROTO_IP)
	}

	log.Errorf("mode %s is not supported", mode)
}

func doProxy(bindProtocol, connectProtocol int) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, bindProtocol)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)

	bindAddr := syscall.SockaddrInet4{
		Port: localPort,
		Addr: [4]byte{0, 0, 0, 0},
	}
	err = syscall.Bind(fd, &bindAddr)
	if err != nil {
		log.Fatal(err)
	}

	err = syscall.Listen(fd, 5)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started to listening...")

	for {
		fd2, rAddr, err := syscall.Accept(fd)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Accepted connection (fd=%d)", fd2)
		go handleConnection(fd2, rAddr.(*syscall.SockaddrInet4), connectProtocol)
	}
}

func handleConnection(fd int, ra *syscall.SockaddrInet4, connectProtocol int) error {
	defer syscall.Close(fd)

	rFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, connectProtocol)
	if err != nil {
		log.Error(err)
		return err
	}
	defer syscall.Close(rFd)

	remoteSockAddr := syscall.SockaddrInet4{
		Port: remotePort,
	}
	copy(remoteSockAddr.Addr[:], remoteAddr.To4())

	err = syscall.Connect(rFd, &remoteSockAddr)
	if err != nil {
		log.Error(err)
		return err
	}

	clientAddr := net.IPv4(ra.Addr[0], ra.Addr[1], ra.Addr[2], ra.Addr[3])
	endpoints := fmt.Sprintf("src=%s:%d dst=%s:%d", clientAddr.String(), ra.Port, remoteAddr.String(), remotePort)
	log.Printf("connected to remote(%s)", endpoints)

	err = copyFdStream(rFd, fd)
	if err != nil {
		log.Error(err)
	}

	log.Printf("proxy finished(%s)", endpoints)
	return nil
}

func copyFdStream(fd1 int, fd2 int) error {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epfd)

	var eventFd1 syscall.EpollEvent
	var eventFd2 syscall.EpollEvent
	var events [10]syscall.EpollEvent

	eventFd1.Events = syscall.EPOLLIN | syscall.EPOLLRDHUP
	eventFd1.Fd = int32(fd1)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd1, &eventFd1)
	if err != nil {
		return err
	}

	eventFd2.Events = syscall.EPOLLIN | syscall.EPOLLRDHUP
	eventFd2.Fd = int32(fd2)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd2, &eventFd2)
	if err != nil {
		return err
	}

	err = syscall.SetNonblock(fd1, true)
	if err != nil {
		return err
	}

	err = syscall.SetNonblock(fd2, true)
	if err != nil {
		return err
	}

	b := make([]byte, 65535)
	for {
		nevents, err := syscall.EpollWait(epfd, events[:], -1)
		if err != nil {
			// goroutine can cause EINTR
			if err == syscall.EINTR {
				continue
			}
			log.Error("EpollWait")
			return err
		}

		close := false
		for ev := 0; ev < nevents; ev++ {
			if events[ev].Events&syscall.EPOLLRDHUP > 0 {
				close = true
			}
		}

		for ev := 0; ev < nevents; ev++ {
			if events[ev].Events&syscall.EPOLLIN == 0 {
				continue
			}
			close = false

			readFd := int(events[ev].Fd)
			var writeFd int
			if readFd == fd1 {
				writeFd = fd2
			} else {
				writeFd = fd1
			}

			readSize, err := syscall.Read(readFd, b)
			if err != nil {
				return err
			}

			if readSize == 0 {
				log.Debugf("READ SIZE == 0")
				return nil
			}

			writeSize, err := syscall.Write(writeFd, b[:readSize])
			if err != nil {
				return err
			}
			log.Debugf("Write(size=%d)", writeSize)
		}

		if close {
			return nil
		}
	}

	// return nil
}
