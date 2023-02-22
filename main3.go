package main

import (
	"go_proxy/ciphers"
	"go_proxy/proxy"
)

func main() {
	//fmt.Println([]byte{})
	//fmt.Println([2]byte{})
	//data := make([]byte, 10)
	//fmt.Println(data)
	//data = make([]byte, 0)
	//fmt.Println(data)
	key := "NrnP0XmA29WRVQiSLhXtKMEnRVA5UyRmq5Qf/RnxIU+zMuZWqSIP3Zheyvsdj50BzfKKd3vOBWnTF2h8i2Pin8By19nhw/fCsQemjTG69XXz8CrJLRsLpxzgDV0wCWpHEtqgToSGYuzQl8bEZZwrAgo4SBqwKdxav7uIQq+ipUvlg5kgiQO+PnhwTVK8jpUEk1hJ1G0AtPxf6mD+QPQUsqMeV/jMZ9LobHo3/7cW7ulhc2R/f\nT+uvVv5+riBqlHfjJuQyz3k3iwO4xO2PLX2pDo1b1k71nQlQZYMghHYhyMGdp7F663nRMiFVC8zoUZDcUxcNMd+ShgmEKxrbqjvmg"
	caesar := ciphers.NewCaesarCipher(key)
	a := proxy.NewSecureSocksProxy("127.0.0.1:8000", "127.0.0.1:20001", caesar)
	//a := proxy.NewSecureHttpProxy("127.0.0.1:8000", "127.0.0.1:20001", caesar)
	a.Start()
}
