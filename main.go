package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"
)

func parseMapPortInfo(data []byte) error {
	fmt.Printf("端口映射信息(remote_port-local_port):%s\n", data)
	return nil
}

func main() {
	localIP := "127.0.0.1"  // 填写想要映射的地址
	serverIP := "127.0.0.1" // 代理服务器地址(外网主机地址)
	serverPort := 3333      // 代理服务器端口(外网主机端口)

	for {
		conn, err := net.Dial("tcp4", fmt.Sprintf("%s:%d", serverIP, serverPort))
		if err != nil {
			time.Sleep(time.Second * 5)
			continue
		}

		// 发送想要导出的端口格式(以下导出本地的22端口和3389端口)):
		// register\r\n$port\r\n$port\r\n...$port\r\n
		conn.Write([]byte("register\r\n22\r\n3389\r\n"))

		r := bufio.NewReader(conn)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		data, _, err := r.ReadLine()
		if err != nil || string(data) != "notify" {
			time.Sleep(time.Second * 5)
			continue
		}

		data, _, err = r.ReadLine()
		if err != nil || parseMapPortInfo(data) != nil {
			time.Sleep(time.Second * 5)
			continue
		}

		conn.SetReadDeadline(time.Now().Add(0xffff * time.Hour))

		for {
			data, _, err := bufio.NewReader(conn).ReadLine()
			if err != nil {
				break
			}

			flieds := strings.Split(string(data), "-")
			if len(flieds) != 3 {
				break
			}

			destConn, err := net.Dial("tcp4", fmt.Sprintf("%s:%s", localIP, flieds[2]))
			if err != nil {
				continue
			}

			srcConn, err := net.Dial("tcp4", fmt.Sprintf("%s:%s", serverIP, flieds[1]))
			if err != nil {
				destConn.Close()
				continue
			}

			srcConn.Write([]byte(fmt.Sprintf("dataer\r\n%s\r\n", flieds[0])))

			trr := io.TeeReader(destConn, srcConn)
			trw := io.TeeReader(srcConn, destConn)

			go func() {
				defer destConn.Close()
				defer srcConn.Close()
				ioutil.ReadAll(trr)
			}()

			go func() {
				defer destConn.Close()
				defer srcConn.Close()
				ioutil.ReadAll(trw)
			}()
		}

		conn.Close()
	}
}
