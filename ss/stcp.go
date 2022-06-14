package ss

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"forward/util"
	"io"
	"net"
	"time"
)

const (
	VERSION              = 0
	TIMESTAMP_MAX_OFFSET = 60
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
				buf := make([]byte, 64)
				_, err = io.ReadFull(sconn, buf) // 64 random
				if err != nil {
					return
				}
				_, err = io.ReadFull(sconn, buf[:9])
				if err != nil {
					return
				}
				if buf[0] != VERSION {
					err = errors.New("invalid version")
					return
				}
				timestamp := binary.BigEndian.Uint64(buf[1:9])
				offset := time.Now().Unix() - int64(timestamp)
				if util.Math.AbsInt64(offset) > TIMESTAMP_MAX_OFFSET {
					err = errors.New("invalid timestamp")
					return
				}
				_, err = io.CopyN(sconn, rand.Reader, 64) // 64 random
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
	io.CopyN(sconn, rand.Reader, 64) // 64 random
	sconn.Write([]byte{VERSION})     // 1 version
	timestamp := make([]byte, 8)     // 8 timestamp
	binary.BigEndian.PutUint64(timestamp, uint64(time.Now().Unix()))
	sconn.Write(timestamp)
	buf := make([]byte, 64)
	_, err = io.ReadFull(sconn, buf) // 64 random from server
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
	s.encrypter.XORKeyStream(b, b)
	return s.Conn.Write(b)
}
