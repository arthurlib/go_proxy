package proxy

import (
	"fmt"
	"go_proxy/ciphers"
	"net"
)

// 协议说明
// https://wiyi.org/socks5-protocol-in-deep.html

type SocksProxy struct {
	listen net.Listener
}

func (p *SocksProxy) Start() {
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

func (p *SocksProxy) Close() {
	_ = p.listen.Close()
}

func (p *SocksProxy) process(conn net.Conn) {

	data := make([]byte, 260)
	// 握手
	length, err := conn.Read(data)
	if err != nil || data[0] != 0x05 {
		_ = conn.Close()
		return
	}
	//fmt.Println(data[:length])
	_, err = conn.Write([]byte{0x05, 0x00}) //无需认证
	if err != nil {
		_ = conn.Close()
		return
	}
	// 请求细节
	length, err = conn.Read(data)
	if err != nil || data[0] != 0x05 || length < 7 || data[1] != 0x01 { // 只处理了 tcp
		_ = conn.Close()
		return
	}
	//fmt.Println(data[:length])

	// 获得端口
	port := BytesToInt16(data[length-2 : length])
	if port == 0 {
		_ = conn.Close()
		return
	}
	//fmt.Println(port)

	var host string
	if data[3] == 0x01 {
		host = string(data[4:8])
	} else if data[3] == 0x03 {
		host = string(data[5 : length-2])
	} else if data[3] == 0x04 {
		host = string(data[4:20])
	} else {
		_ = conn.Close()
		return
	}
	//fmt.Println(host)
	//fmt.Println(host + ":" + strconv.Itoa(port))

	//addrs, err := net.LookupIP(host)
	//fmt.Println(addrs)
	//fmt.Println(addrs[0].String() + ":" + strconv.Itoa(port))
	//remoteConn, err := net.Dial("tcp", addrs[0].String()+":"+strconv.Itoa(port))
	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		//fmt.Printf("conn server failed, err:%v\n", err)
		_ = conn.Close()
		return
	}
	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		_ = conn.Close()
		return
	}
	go ReadAndWrite(conn, remoteConn)
	go ReadAndWrite(remoteConn, conn)
}

func NewSocksProxy(address string) *SocksProxy {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", address) // "127.0.0.1:8000
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return nil
	}
	return &SocksProxy{listen: listen}
}

type SecureSocksProxy struct {
	*SocksProxy
	cipher    ciphers.Cipher
	proxyAddr string
}

func (p *SecureSocksProxy) Start() {
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

func (p *SecureSocksProxy) process(conn net.Conn) {

	data := make([]byte, 260)
	// 握手
	length, err := conn.Read(data)
	if err != nil || data[0] != 0x05 {
		_ = conn.Close()
		return
	}
	//fmt.Println(data[:length])
	_, err = conn.Write([]byte{0x05, 0x00}) //无需认证
	if err != nil {
		_ = conn.Close()
		return
	}
	// 请求细节
	length, err = conn.Read(data)
	if err != nil || data[0] != 0x05 || length < 7 || data[1] != 0x01 { // 只处理了 tcp
		_ = conn.Close()
		return
	}
	//fmt.Println(data[:length])

	// 获得端口
	port := BytesToInt16(data[length-2 : length])
	if port == 0 {
		_ = conn.Close()
		return
	}
	//fmt.Println(port)

	var host []byte
	if data[3] == 0x01 {
		host = data[4:8]
	} else if data[3] == 0x03 {
		host = data[5 : length-2]
	} else if data[3] == 0x04 {
		host = data[4:20]
	} else {
		_ = conn.Close()
		return
	}
	//fmt.Println(host)
	//fmt.Println(host + ":" + strconv.Itoa(port))

	//addrs, err := net.LookupIP(host)
	//fmt.Println(addrs)
	//fmt.Println(addrs[0].String() + ":" + strconv.Itoa(port))
	//remoteConn, err := net.Dial("tcp", addrs[0].String()+":"+strconv.Itoa(port))
	//remoteConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	remoteConn, err := NewSecureClient(p.proxyAddr, p.cipher)
	if err != nil {
		//fmt.Printf("conn server failed, err:%v\n", err)
		_ = conn.Close()
		return
	}
	err = remoteConn.RequestDestAddr(host, data[length-2:length])
	if err != nil {
		_ = remoteConn.Close()
		_ = conn.Close()
		return
	}

	_, err = conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		_ = remoteConn.Close()
		_ = conn.Close()
		return
	}
	go ReadAndWrite(conn, remoteConn)
	go ReadAndWrite(remoteConn, conn)
}

func NewSecureSocksProxy(addr, proxyAddr string, cipher ciphers.Cipher) *SecureSocksProxy {
	// 建立 tcp 服务
	listen, err := net.Listen("tcp", addr) // "127.0.0.1:8000
	if err != nil {
		fmt.Printf("listen failed, err:%v\n", err)
		return nil
	}
	return &SecureSocksProxy{&SocksProxy{listen: listen}, cipher, proxyAddr}
}
