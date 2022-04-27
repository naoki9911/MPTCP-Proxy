package main

import (
	"net"
	"strconv"
	"strings"
	"sync"
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
		go handleConnection(fd2, rAddr, connectProtocol)
	}
}

func handleConnection(fd int, ra syscall.Sockaddr, connectProtocol int) error {
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

	log.Printf("connected to remote server(addr=%s port=%d)", remoteAddr.String(), remotePort)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := copyFdStream(fd, rFd, "fd -> rFd")
		if err != nil {
			log.Error(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := copyFdStream(rFd, fd, "rFd -> fd")
		if err != nil {
			log.Error(err)
		}
	}()

	wg.Wait()

	log.Printf("proxy finised")
	return nil
}

func copyFdStream(fd1 int, fd2 int, logPrefix string) error {
	b := make([]byte, 65535)
	for {
		readSize, err := syscall.Read(fd1, b)
		if err != nil {
			return err
		}

		if readSize == 0 {
			return nil
		}

		writeSize, err := syscall.Write(fd2, b[:readSize])
		if err != nil {
			return err
		}
		log.Debugf("%s Write(size=%d)", logPrefix, writeSize)
	}

	// return nil
}
