package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"net"

	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

// Server represents a honeypot SSH server. The server masquerades as an open SSH server
// and allows any username/password combination to login. Interesting activity is logged
// and sent back on a Go channel.
type Server struct {
	// Address is the combination of host:port to listen on. If port is empty a random port
	// will be chosen.
	Address string

	ioTimeout time.Duration

	// quit is a channel that signals the server to shutdown
	quit chan bool
}

// NewServer creates an SSH honeypot server capable of serving connections.
func NewServer(address string) (*Server, error) {
	quit := make(chan bool)
	timeout := 5 * time.Second

	return &Server{address, timeout, quit}, nil
}

// Serve starts the honeypot SSH server. It returns a channel of Connection pointers,
// errors while processing the SSH connection and a error. If the server fails to start,
// the last error will contain the reason.
func (s Server) Serve() (chan *Connection, chan error, error) {
	config := &ssh.ServerConfig{
		// Allow everyone to login. This is a honeypot 😀
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			m := make(map[string]string)
			sessionID := base64.StdEncoding.EncodeToString(c.SessionID())
			m[sessionID] = string(pass)
			perms := &ssh.Permissions{
				Extensions: m, // Use extensions as a mechanism to save the session ID -> password mapping
			}
			return perms, nil
		},
	}

	hostKey, err := hostKey()
	if err != nil {
		return nil, nil, errors.Wrap(err, "host key generation failed")
	}

	config.AddHostKey(hostKey)

	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, nil, errors.Wrap(err, "tcp listen failed")
	}

	conns := make(chan *Connection)
	connErrs := make(chan error)

	go s.accept(listener, config, conns, connErrs)

	return conns, connErrs, nil
}

// Stop signals the server to stop serving requests and shutdown.
func (s *Server) Stop() {
	s.quit <- true
}

func (s Server) accept(listener net.Listener, config *ssh.ServerConfig,
	connections chan *Connection, errs chan error) {
	defer func() {
		close(connections)
		close(errs)
	}()

	for {
		select {
		case <-s.quit:
			return
		default:
			conn, err := listener.Accept()

			conn.SetDeadline(time.Now().Add(s.ioTimeout))

			if err != nil {
				errs <- errors.Wrapf(err, "accept connection %v failed", conn)
				continue
			}
			serverConn, _, reqs, err := ssh.NewServerConn(conn, config)
			if err != nil {
				errs <- errors.Wrap(err, "ssh handshake failed")
				continue
			}
			connection := newConnection(serverConn)
			go ssh.DiscardRequests(reqs) //TODO: process requests
			connections <- connection
		}
	}
}

func hostKey() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}
	return signer, nil
}
