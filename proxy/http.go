package proxy

import (
	"fmt"
	"go_proxy/ciphers"
	"net"
	"strconv"
	"strings"
)

type HttpProxy struct {
	listen net.Listener
}

func (p *HttpProxy) Start() {
	for {
		// 等待客户端建立连接
		conn, err := p.listen.Accept()
		if err != nil {
			break
		}
		// 启动一个单独的 goroutine 去处理连接
		go p.process(conn)
	}
}

func (p *HttpProxy) Close() {
	_ = p.listen.Close()
}

func (p *HttpProxy) process(conn net.Conn) {

	data := make([]byte, 1024)
	// 握手
	length, err := conn.Read(data)
	if err != nil {
		_ = conn.Close()
		return
	}
	var host string
	var port int
	for _, line := range strings.Split(string(data[:length]), "\r\n") {
		if strings.HasPrefix(line, "Host") {
			addr := line[6:]
			if strings.Contains(addr, ":") {
				remote := strings.Split(addr, ":")
				host = remote[0]
				port, _ = strconv.Atoi(remote[1])
			} else {
				host = addr
				port = 80
			}
			break
		}
	}
	//fmt.Println(host + ":" + port)
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		//fmt.Printf("conn server failed, err:%v\n", err)
		_ = conn.Close()
		return
	}

	if port == 443 {
		_, err = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			_ = conn.Close()
			return
		}
	} else {
		_, err = remoteConn.Write(data[:length])
		if err != nil {
			_ = conn.Close()
			return
		}
	}

	go ReadAndWrite(conn, remoteConn)
	go ReadAndWrite(remoteConn, conn)
}

func NewHttpProxy(address string) *HttpProxy {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", address) // "127.0.0.1:8000
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return nil
	}
	return &HttpProxy{listen: listen}
}

type SecureHttpProxy struct {
	*HttpProxy
	cipher    ciphers.Cipher
	proxyAddr string
}

func (p *SecureHttpProxy) Start() {
	for {
		// 等待客户端建立连接
		conn, err := p.listen.Accept()
		if err != nil {
			break
		}
		// 启动一个单独的 goroutine 去处理连接
		go p.process(conn)
	}
}

func (p *SecureHttpProxy) process(conn net.Conn) {

	data := make([]byte, 1024)
	// 握手
	length, err := conn.Read(data)
	if err != nil {
		_ = conn.Close()
		return
	}
	var host string
	var port int
	for _, line := range strings.Split(string(data[:length]), "\r\n") {
		if strings.HasPrefix(line, "Host") {
			addr := line[6:]
			if strings.Contains(addr, ":") {
				remote := strings.Split(addr, ":")
				host = remote[0]
				port, _ = strconv.Atoi(remote[1])
			} else {
				host = addr
				port = 80
			}
			break
		}
	}
	//fmt.Println(host + ":" + port)
	remoteConn, err := NewSecureClient(p.proxyAddr, p.cipher)
	if err != nil {
		//fmt.Printf("conn server failed, err:%v\n", err)
		_ = conn.Close()
		return
	}
	err = remoteConn.RequestDestAddr([]byte(host), Int16ToBytes(int16(port)))
	if err != nil {
		_ = remoteConn.Close()
		_ = conn.Close()
		return
	}

	if port == 443 {
		_, err = conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
		if err != nil {
			_ = conn.Close()
			return
		}
	} else {
		_, err = remoteConn.Write(data[:length])
		if err != nil {
			_ = conn.Close()
			return
		}
	}

	go ReadAndWrite(conn, remoteConn)
	go ReadAndWrite(remoteConn, conn)
}

func NewSecureHttpProxy(address, proxyAddr string, cipher ciphers.Cipher) *SecureHttpProxy {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", address) // "127.0.0.1:8000
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return nil
	}
	return &SecureHttpProxy{&HttpProxy{listen: listen}, cipher, proxyAddr}
}
