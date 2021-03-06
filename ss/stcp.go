package ss

import (
	"crypto/rand"
	"errors"
	"forward/util"
	"io"
	"net"
	"time"
)

type STCPListener struct {
	net.Listener
	secret []byte
	c      chan *acceptPackage
}

type STCPConn struct {
	net.Conn
	encrypter *util.StreamEncrypter
	decrypter *util.StreamDecrypter
}

type acceptPackage struct {
	sconn *STCPConn
	err   error
}

func ListenSTCP(network string, laddr *net.TCPAddr, secret []byte) (*STCPListener, error) {
	l, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}
	sl := &STCPListener{
		Listener: l,
		secret:   secret,
		c:        make(chan *acceptPackage),
	}
	go func(sl *STCPListener) {
		for {
			conn, err := sl.Listener.Accept()
			if err != nil {
				sl.c <- &acceptPackage{
					sconn: nil,
					err:   err,
				}
				continue
			}
			go func(sl *STCPListener, conn net.Conn) {
				var sconn = &STCPConn{
					Conn:      conn,
					encrypter: util.StreamCipher.NewStreamEncrypter(sl.secret),
					decrypter: util.StreamCipher.NewStreamDecrypter(sl.secret),
				}
				var err error
				defer func() {
					if err != nil {
						conn.Close()
					}
					sl.c <- &acceptPackage{
						sconn: sconn,
						err:   err,
					}
				}()
				/* handshake */
				sconn.SetDeadline(time.Now().Add(3 * time.Second))
				defer sconn.SetDeadline(time.Time{})
				_, err = io.CopyN(io.Discard, sconn, 32)
				if err != nil {
					return
				}
				buf := make([]byte, 64)
				rand.Read(buf[32:])
				_, err = sconn.Write(buf[32:])
				if err != nil {
					return
				}
				_, err = io.ReadFull(sconn, buf[:32])
				if err != nil {
					return
				}
				for i := 0; i < 32; i++ {
					if buf[i] != buf[i+32] {
						err = errors.New("invaild random")
						return
					}
				}
				_, err = io.CopyN(sconn, rand.Reader, 32)
				if err != nil {
					return
				}
				/* handshake */
			}(sl, conn)
		}
	}(sl)
	return sl, nil
}

func (sl *STCPListener) Accept() (*STCPConn, error) {
	ap := <-sl.c
	if ap.err != nil {
		return nil, ap.err
	}
	return ap.sconn, nil
}

func DialSTCP(network string, laddr *net.TCPAddr, raddr *net.TCPAddr, secret []byte) (*STCPConn, error) {
	conn, err := net.DialTCP(network, laddr, raddr)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()
	sconn := &STCPConn{
		Conn:      conn,
		encrypter: util.StreamCipher.NewStreamEncrypter(secret),
		decrypter: util.StreamCipher.NewStreamDecrypter(secret),
	}
	/* handshake */
	sconn.SetDeadline(time.Now().Add(3 * time.Second))
	defer sconn.SetDeadline(time.Time{})
	_, err = io.CopyN(sconn, rand.Reader, 32)
	if err != nil {
		return nil, err
	}
	_, err = io.CopyN(sconn, sconn, 32)
	if err != nil {
		return nil, err
	}
	_, err = io.CopyN(io.Discard, sconn, 32)
	if err != nil {
		return nil, err
	}
	/* handshake */
	return sconn, nil
}

func (s *STCPConn) Read(b []byte) (int, error) {
	n, err := s.Conn.Read(b)
	if err != nil {
		return n, err
	}
	s.decrypter.XORKeyStream(b[:n], b[:n])
	return n, nil
}

func (s *STCPConn) Write(b []byte) (int, error) {
	dst := make([]byte, len(b))
	s.encrypter.XORKeyStream(dst, b)
	return s.Conn.Write(dst)
}
