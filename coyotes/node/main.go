package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/mylxsw/coyotes/log"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "coyotes"
	app.Version = "2.0"
	app.Usage = "一款轻量级的分布式任务执行队列"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "mylxsw",
			Email: "mylxsw@aicode.cc",
		},
	}

	app.Flags = []cli.Flag{}

	app.Action = func(c *cli.Context) error {
		conn, err := net.Dial("tcp", "127.0.0.1:10086")
		if err != nil {
			log.Error("连接到服务器失败:%v", err)
			return err
		}
		defer conn.Close()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			buffer := make([]byte, 1024)
			for {
				conn.Read(buffer)
				fmt.Println(string(buffer))
				time.Sleep(2 * time.Second)
			}
		}()

		go func() {
			defer wg.Done()
			for {
				conn.Write([]byte("Hello from client"))
				time.Sleep(2 * time.Second)
			}
		}()

		wg.Wait()

		return nil
	}

	app.Run(os.Args)
}
