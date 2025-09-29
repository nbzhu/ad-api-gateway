package initialize

import (
	"fmt"
	pb "github.com/nbzhu/ad-api-gateway-proto"
	"github.com/nbzhu/ad-api-gateway/api/ocean"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
)

func InitServer(port int) {
	go func() {
		log.Println(http.ListenAndServe("localhost:50052", nil))
	}()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(UnaryInterceptor))
	pb.RegisterApiServer(grpcServer, &ocean.Api{})

	log.Println("gRPC 服务启动，监听端口 :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
