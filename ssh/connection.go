package ssh

import (
	"encoding/base64"
	"net"

	"golang.org/x/crypto/ssh"
)

// Connection represents an SSH connection made by an attacker. It contains
// information about the origin of the attack and payloads attempted by the attacker.
type Connection struct {
	SessionID      string
	SourceIP       string
	SourceHostName string
	UserName       string
	Password       string
	ClientVersion  string
	Payloads       []interface{}
}

func newConnection(serverConn *ssh.ServerConn) *Connection {
	sourceIP := serverConn.RemoteAddr().String()
	sessionID := base64.StdEncoding.EncodeToString(serverConn.SessionID())
	c := &Connection{
		SessionID:     sessionID,
		SourceIP:      sourceIP,
		Password:      serverConn.Permissions.Extensions[sessionID], // We use extensions to store auth info from callback
		ClientVersion: string(serverConn.ClientVersion()),
		UserName:      serverConn.User(),
	}
	hostName, err := net.LookupAddr(sourceIP)
	if err != nil {
		// TODO: system error logging
	} else {
		c.SourceHostName = hostName[0]
	}
	return c
}
