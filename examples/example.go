package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func main() {
	var err error
	var ftpConn *common.EkpsFtp

	addr := net.JoinHostPort("192.168.10.1", strconv.Itoa(21))
	if ftpConn, err = common.FtpConnect(addr, time.Duration(5)*time.Second); err != nil {
		fmt.Println(err.Error())
	}
	defer ftpConn.Quit()

	if err := ftpConn.Login("userid", "password"); err != nil {
		fmt.Println(err.Error())
		return
	}

	ftpConn.SetPasv(true)

	fileList, err := ftpConn.List("/home/kzdev")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	for _, file := range fileList {
		fmt.Println(file)
	}
}
