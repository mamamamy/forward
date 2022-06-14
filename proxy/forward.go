package proxy

import (
	"forward/ss"
	"io"
	"log"
	"net"
)

type ForwardServer struct {
	lnetwork string
	rnetwork string
	laddr    *net.TCPAddr
	raddr    *net.TCPAddr
	secret   []byte
}

type ForwardClient struct {
	lnetwork string
	rnetwork string
	laddr    *net.TCPAddr
	raddr    *net.TCPAddr
	secret   []byte
}

func NewForwardServer(lnetwork, rnetwork, laddress, raddress string, secret []byte) (*ForwardServer, error) {
	laddr, err := net.ResolveTCPAddr(lnetwork, laddress)
	if err != nil {
		return nil, err
	}
	raddr, err := net.ResolveTCPAddr(rnetwork, raddress)
	if err != nil {
		return nil, err
	}
	return &ForwardServer{
		lnetwork: lnetwork,
		rnetwork: rnetwork,
		laddr:    laddr,
		raddr:    raddr,
		secret:   secret,
	}, nil
}

func NewForwardClient(lnetwork, rnetwork, laddress, raddress string, secret []byte) (*ForwardClient, error) {
	laddr, err := net.ResolveTCPAddr(lnetwork, laddress)
	if err != nil {
		return nil, err
	}
	raddr, err := net.ResolveTCPAddr(rnetwork, raddress)
	if err != nil {
		return nil, err
	}
	return &ForwardClient{
		lnetwork: lnetwork,
		rnetwork: rnetwork,
		laddr:    laddr,
		raddr:    raddr,
		secret:   secret,
	}, nil
}

func relay(dst, src net.Conn) {
	defer src.Close()
	defer dst.Close()
	defer log.Println(src.RemoteAddr(), "->", dst.RemoteAddr(), "close")
	io.Copy(dst, src)
}

func (fs *ForwardServer) Run() {
	sl, err := ss.ListenSTCP("tcp", fs.laddr, fs.secret)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("server listen at", sl.Addr())
	for {
		sconn, err := sl.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(sconn net.Conn) {
			conn, err := net.DialTCP("tcp", &net.TCPAddr{}, fs.raddr)
			if err != nil {
				sconn.Close()
				return
			}
			go relay(conn, sconn)
			go relay(sconn, conn)
		}(sconn)
	}
}

func (fc *ForwardClient) Run() {
	l, err := net.ListenTCP("tcp", fc.laddr)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("client listen at", l.Addr())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func(conn net.Conn) {
			sconn, err := ss.DialSTCP("tcp", &net.TCPAddr{}, fc.raddr, fc.secret)
			if err != nil {
				conn.Close()
				return
			}
			go relay(sconn, conn)
			go relay(conn, sconn)
		}(conn)
	}
}
