package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"go_proxy/ciphers"
	"net"
)

var (
	firstMsg     = "hei bro, i need help"
	firstMsgResp = "easy, tell me"
)

type SecureServer struct {
	listen net.Listener
	cipher ciphers.Cipher
}

func (s *SecureServer) Start() {
	for {
		// 等待客户端建立连接
		conn, err := s.listen.Accept()
		if err != nil {
			break
		}
		// 启动一个单独的 goroutine 去处理连接
		go s.process(&EncryptedStream{conn, s.cipher})
	}
}

func (s *SecureServer) Close() {
	_ = s.listen.Close()
}

func (s *SecureServer) process(conn Stream) {
	data := make([]byte, 50)
	length, err := conn.Read(data)
	if err != nil {
		_ = conn.Close()
		return
	}
	if string(data[:length]) != firstMsg {
		_, _ = conn.Write([]byte("fuck you"))
		_ = conn.Close()
		return
	}

	_, err = conn.Write([]byte(firstMsgResp))
	if err != nil {
		_ = conn.Close()
		return
	}

	// 请求细节
	length, err = conn.Read(data)
	if err != nil {
		_ = conn.Close()
		return
	}
	host := string(data[:length-2])
	port := BytesToInt16(data[length-2 : length]) // 获得端口
	if port == 0 {
		_ = conn.Close()
		return
	}
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		_, err = conn.Write([]byte{0x01}) // 连接失败
		_ = conn.Close()
		return
	}
	_, err = conn.Write([]byte{0x00})
	if err != nil {
		_ = conn.Close()
		return
	}
	go ReadAndWrite(conn, remoteConn)
	go ReadAndWrite(remoteConn, conn)
}

// NewSecureServer 只处理了 tcp
func NewSecureServer(address string, cipher ciphers.Cipher) *SecureServer {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", address) // "127.0.0.1:8000
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return nil
	}
	return &SecureServer{listen, cipher}
}

type SecureClient struct {
	*EncryptedStream
}

func (c *SecureClient) auth() error {
	_, err := c.Write([]byte(firstMsg))
	if err != nil {
		return err
	}

	data := make([]byte, 30)
	length, err := c.Read(data)
	if err != nil {
		_ = c.Close()
		return err
	}
	if string(data[:length]) != firstMsgResp {
		_ = c.Close()
		return errors.New("auth error")
	}
	return nil
}

func (c *SecureClient) RequestDestAddr(host, port []byte) error {
	_, err := c.Write(bytes.Join([][]byte{host, port}, []byte{}))
	if err != nil {
		return err
	}

	data := make([]byte, 5)
	_, err = c.Read(data)
	if err != nil {
		return err
	}
	if data[0] != 0x00 {
		return errors.New("connect error")
	}
	return nil
}

func NewSecureClient(address string, cipher ciphers.Cipher) (*SecureClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		//fmt.Printf("conn server failed, err:%v\n", err)
		return nil, err
	}
	transportClient := &SecureClient{&EncryptedStream{conn, cipher}}
	err = transportClient.auth()
	if err != nil {
		_ = transportClient.Close()
		return nil, err
	}
	return transportClient, nil
}
