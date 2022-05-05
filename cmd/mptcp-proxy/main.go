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
const BUF_SIZE = 65536

var (
	remoteAddr   net.IP
	remotePort   int
	localPort    int
	mode         string
	transparent  bool
	disableMPTCP bool
)

func main() {
	if log.GetLevel() == log.DebugLevel {
		log.SetReportCaller(true)
	}

	var err error

	flag.BoolVarP(&transparent, "transparent", "t", false, "Enable transparent mode")
	flag.StringVarP(&mode, "mode", "m", "", "specify mode (server or client)")
	flag.IntVarP(&localPort, "port", "p", 0, "local bind port")
	flag.BoolVar(&disableMPTCP, "disable-mptcp", false, "Disable MPTCP")
	var rAddr *string = flag.StringP("remote", "r", "", "remote address (ex. 127.0.0.1:8080)")
	flag.Parse()

	if localPort == 0 || mode == "" {
		flag.Usage()
		return
	}

	if mode != "server" && mode != "client" {
		flag.Usage()
		return
	}

	if !transparent {
		if *rAddr == "" {
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
	}

	log.Infof("starting proxy...")
	if disableMPTCP {
		doProxy(syscall.IPPROTO_IP, syscall.IPPROTO_IP)
	} else {
		if mode == "client" {
			doProxy(syscall.IPPROTO_IP, IPPROTO_MPTCP)
		} else if mode == "server" {
			doProxy(IPPROTO_MPTCP, syscall.IPPROTO_IP)
		}
	}

	log.Errorf("mode %s is not supported", mode)
}

const SO_ORIGINAL_DST = 80

func getOriginalDestination(sockfd int) (net.IP, int, error) {
	// this code is copied from https://gist.github.com/fangdingjun/11e5d63abe9284dc0255a574a76bbcb1

	// Get original destination
	// this is the only syscall in the Golang libs that I can find that returns 16 bytes
	// Example result: &{Multiaddr:[2 0 31 144 206 190 36 45 0 0 0 0 0 0 0 0] Interface:0}
	// port starts at the 3rd byte and is 2 bytes long (31 144 = port 8080)
	// IPv6 version, didn't find a way to detect network family
	//addr, err := syscall.GetsockoptIPv6Mreq(int(clientConnFile.Fd()), syscall.IPPROTO_IPV6, IP6T_SO_ORIGINAL_DST)
	// IPv4 address starts at the 5th byte, 4 bytes long (206 190 36 45)
	addr, err := syscall.GetsockoptIPv6Mreq(sockfd, syscall.IPPROTO_IP, SO_ORIGINAL_DST)
	log.Debugf("getOriginalDst(): SO_ORIGINAL_DST=%+v\n", addr)
	if err != nil {
		return nil, 0, err
	}

	ip := net.IPv4(addr.Multiaddr[4], addr.Multiaddr[5], addr.Multiaddr[6], addr.Multiaddr[7])
	port := int(addr.Multiaddr[2])<<8 | int(addr.Multiaddr[3])

	return ip, port, nil
}

func doProxy(bindProtocol, connectProtocol int) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, bindProtocol)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(fd)

	err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		log.Fatal(err)
	}

	bindAddr := syscall.SockaddrInet4{
		Port: localPort,
	}

	if transparent {
		bindAddr.Addr = [4]byte{127, 0, 0, 1}
		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_TRANSPARENT, 1)
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		bindAddr.Addr = [4]byte{0, 0, 0, 0}
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

		remoteSockAddr := &syscall.SockaddrInet4{}
		if transparent {
			ip, port, err := getOriginalDestination(fd2)
			if err != nil {
				log.Printf("failed to get original destination %s", err)
				continue
			}
			copy(remoteSockAddr.Addr[:], ip.To4())
			remoteSockAddr.Port = port
		} else {
			copy(remoteSockAddr.Addr[:], remoteAddr.To4())
			remoteSockAddr.Port = remotePort
		}

		go handleConnection(fd2, rAddr.(*syscall.SockaddrInet4), remoteSockAddr, connectProtocol)
	}
}

func handleConnection(fd int, src, dst *syscall.SockaddrInet4, connectProtocol int) error {
	defer syscall.Close(fd)

	rFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, connectProtocol)
	if err != nil {
		log.Error(err)
		return err
	}
	defer syscall.Close(rFd)

	if connectProtocol == syscall.IPPROTO_IP {
		err = syscall.SetsockoptInt(rFd, syscall.SOL_TCP, syscall.TCP_NODELAY, 1)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		err = syscall.SetsockoptInt(fd, syscall.SOL_TCP, syscall.TCP_NODELAY, 1)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	err = syscall.Connect(rFd, dst)
	if err != nil {
		log.Error(err)
		return err
	}

	srcAddr := net.IPv4(src.Addr[0], src.Addr[1], src.Addr[2], src.Addr[3])
	dstAddr := net.IPv4(dst.Addr[0], dst.Addr[1], dst.Addr[2], dst.Addr[3])
	endpoints := fmt.Sprintf("src=%s:%d dst=%s:%d", srcAddr.String(), src.Port, dstAddr.String(), dst.Port)
	log.Printf("connected to remote(%s)", endpoints)

	//log.SetLevel(log.DebugLevel)
	err = copyFdStream(fd, rFd)
	if err != nil {
		log.Error(err)
	}

	log.Printf("proxy finished(%s)", endpoints)
	return nil
}

func isEpollEventFlagged(events []syscall.EpollEvent, fd int, flag int) bool {
	for _, event := range events {
		if int(event.Fd) != fd {
			continue
		}

		if int(event.Events)&flag > 0 {
			return true
		} else {
			return false
		}
	}

	return false
}

func copyFdStream(fd1 int, fd2 int) error {
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epfd)

	epWritefd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epWritefd)

	var eventFd1 syscall.EpollEvent
	var eventFd2 syscall.EpollEvent
	var eventWriteFd1 syscall.EpollEvent
	var eventWriteFd2 syscall.EpollEvent

	events := make([]syscall.EpollEvent, 10)
	eventsWrite := make([]syscall.EpollEvent, 10)

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

	eventWriteFd1.Events = syscall.EPOLLOUT
	eventWriteFd1.Fd = int32(fd1)
	err = syscall.EpollCtl(epWritefd, syscall.EPOLL_CTL_ADD, fd1, &eventWriteFd1)
	if err != nil {
		return err
	}

	eventWriteFd2.Events = syscall.EPOLLOUT
	eventWriteFd2.Fd = int32(fd2)
	err = syscall.EpollCtl(epWritefd, syscall.EPOLL_CTL_ADD, fd2, &eventWriteFd2)
	if err != nil {
		return err
	}

	b := make([]byte, BUF_SIZE)
	for {
		nevents, err := syscall.EpollWait(epfd, events, -1)
		if err != nil {
			// goroutine can cause EINTR
			if err == syscall.EINTR {
				continue
			}
			log.Error("EpollWait")
			return err
		}
		waitEvents := events[:nevents]

		nevents, err = syscall.EpollWait(epWritefd, eventsWrite, 0)
		if err != nil {
			// goroutine can cause EINTR
			if err == syscall.EINTR {
				continue
			}
			log.Error("EpollWait")
			return err
		}
		eventsWritable := eventsWrite[:nevents]

		close := isEpollEventFlagged(waitEvents, fd1, syscall.EPOLLRDHUP)
		close = close || isEpollEventFlagged(waitEvents, fd2, syscall.EPOLLRDHUP)

		fds := []int{fd1, fd2}
		for fdIdx := range fds {
			logPrefix := fmt.Sprintf("fd%d -> fd%d", fdIdx+1, (fdIdx+2)%(len(fds)+1))
			readFd := fds[fdIdx]
			writeFd := fds[(fdIdx+1)%len(fds)]

			if !isEpollEventFlagged(waitEvents, readFd, syscall.EPOLLIN) {
				continue
			}
			close = false

			if !isEpollEventFlagged(eventsWritable, writeFd, syscall.EPOLLOUT) {
				continue
			}

			readSize, _, err := syscall.Recvfrom(readFd, b, syscall.MSG_DONTWAIT)
			if err != nil {
				log.Errorf("Read")
				return err
			}

			if readSize == 0 {
				log.Debugf("READ SIZE == 0")
				return nil
			}
			log.Debugf("%s Read(size=%d)", logPrefix, readSize)

			writeSize, err := syscall.Write(writeFd, b[:readSize])
			if err != nil {
				log.Errorf("%s Write", logPrefix)
				return err
			}
			log.Debugf("%s Write(size=%d)", logPrefix, writeSize)
		}

		if close {
			return nil
		}
	}

	// return nil
}
