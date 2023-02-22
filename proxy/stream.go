package proxy

import (
	"bytes"
	"encoding/binary"
	"go_proxy/ciphers"
	"net"
)

var (
	blockSize = 512
)

type Stream interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

type EncryptedStream struct {
	net.Conn
	cipher ciphers.Cipher
}

func (s *EncryptedStream) Close() error {
	return s.Conn.Close()
}

func (s *EncryptedStream) Read(b []byte) (n int, err error) {
	if s.cipher != nil {
		dataLenByte := make([]byte, 2)
		_, err = s.Conn.Read(dataLenByte)
		if err != nil {
			return
		}
		dataLen := BytesToInt16(dataLenByte)
		_, err = s.Conn.Read(b[:dataLen])
		if err != nil {
			return
		}
		data := b[:dataLen]
		if s.cipher != nil {
			data = s.cipher.Decrypt(data)
		}
		copy(b, data)
		return len(data), nil

	} else {
		return s.Conn.Read(b)
	}
}

func (s *EncryptedStream) Write(b []byte) (n int, err error) {
	if s.cipher != nil {
		n = len(b)
		num := (n + blockSize - 1) / blockSize // 向上取整
		for i := 0; i < num; i++ {
			start := i * blockSize
			end := (i + 1) * blockSize
			if end > n {
				end = n
			}
			data := b[start:end]
			if s.cipher != nil {
				data = s.cipher.Encrypt(data)
			}
			dataLen := Int16ToBytes(int16(len(data)))
			_, err = s.Conn.Write(bytes.Join([][]byte{dataLen, data}, []byte{}))
			if err != nil {
				n = 0
				return
			}
		}
		return
	} else {
		return s.Conn.Write(b)
	}
}

func ReadAndWrite(inputConn, outputConn Stream) {
	defer func() {
		inputConn.Close()
		outputConn.Close()
	}()
	data := make([]byte, 1024)
	for {
		length, err := inputConn.Read(data)
		if err != nil || length == 0 {
			//fmt.Println(err)
			return
		}
		//fmt.Println(data[:length])
		_, err = outputConn.Write(data[:length])
		if err != nil {
			//fmt.Println(err)
			return
		}
	}
}

func Int16ToBytes(n int16) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, n)
	return bytesBuffer.Bytes()
}

func BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int16
	err := binary.Read(bytesBuffer, binary.BigEndian, &x)
	if err != nil {
		return 0
	}
	return x
}
