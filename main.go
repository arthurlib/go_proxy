package main

import (
	"fmt"
	"go_proxy/ciphers"
)

func main() {
	//fmt.Println([]byte{})
	//fmt.Println([2]byte{})
	//data := make([]byte, 10)
	//fmt.Println(data)
	//data = make([]byte, 0)
	//fmt.Println(data)

	//a := proxy.NewHttpProxy("127.0.0.1:8000")
	//a.Start()

	key := ciphers.RandomPassword()
	fmt.Println(key)
	fmt.Println(string(key))
	fmt.Println(ciphers.DumpsPassword(key))
	fmt.Println(ciphers.LoadsPassword(ciphers.DumpsPassword(key)))

	pp := make([]byte, 100)
	fmt.Println(pp)

	ppp := pp[:50]
	ppp[0] = 0x01
	fmt.Println(pp)
}
