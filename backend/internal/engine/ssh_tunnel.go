package engine

import (
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

// SSHTunnel 通过 SSH 端口转发把远程数据库端口暴露到本机回环地址。
// 客户端代码连接 127.0.0.1:localPort 即可，零侵入桥接任意 TCP 协议。
type SSHTunnel struct {
	localPort  int
	remoteHost string
	remotePort int
	sshClient  *ssh.Client
	listener   net.Listener
}

// NewSSHTunnel 建立到 sshHost:sshPort 的 SSH 连接，并把 remoteHost:remotePort 转发到
// 本地回环 127.0.0.1:localPort。auth 优先尝试 privateKey，其次 password。
// 注意：HostKeyCallback 使用 InsecureIgnoreHostKey — 生产环境应替换为 known_hosts 校验。
func NewSSHTunnel(sshHost string, sshPort int, sshUser, sshPassword string, privateKey string,
	remoteHost string, remotePort int, localPort int) (*SSHTunnel, error) {

	var authMethods []ssh.AuthMethod
	if privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(privateKey))
		if err != nil {
			return nil, fmt.Errorf("parse private key failed: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if sshPassword != "" {
		authMethods = append(authMethods, ssh.Password(sshPassword))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("at least one auth method (privateKey or password) is required")
	}

	config := &ssh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: 生产环境应配置 known_hosts
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial failed: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("local listen failed: %w", err)
	}

	tunnel := &SSHTunnel{
		localPort:  localPort,
		remoteHost: remoteHost,
		remotePort: remotePort,
		sshClient:  client,
		listener:   listener,
	}

	go tunnel.serve()
	return tunnel, nil
}

func (t *SSHTunnel) serve() {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			return
		}
		go t.forward(localConn)
	}
}

func (t *SSHTunnel) forward(localConn net.Conn) {
	defer localConn.Close()
	remoteConn, err := t.sshClient.Dial("tcp", fmt.Sprintf("%s:%d", t.remoteHost, t.remotePort))
	if err != nil {
		return
	}
	defer remoteConn.Close()

	go func() { _, _ = io.Copy(remoteConn, localConn) }()
	_, _ = io.Copy(localConn, remoteConn)
}

// LocalAddr 返回客户端应使用的地址。如果 localPort=0，会用 listener 实际分配的端口。
func (t *SSHTunnel) LocalAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", t.LocalPort())
}

// LocalPort 返回实际监听的本地端口（如果 localPort=0 时由系统分配）。
func (t *SSHTunnel) LocalPort() int {
	if t.listener == nil {
		return t.localPort
	}
	return t.listener.Addr().(*net.TCPAddr).Port
}

func (t *SSHTunnel) Close() error {
	if t.listener != nil {
		_ = t.listener.Close()
	}
	if t.sshClient != nil {
		err := t.sshClient.Close()
		return err
	}
	return nil
}
