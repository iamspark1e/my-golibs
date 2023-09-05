## Kick Start

### Reference

- [一个简单的gRPC项目](https://github.com/Win-Man/NormalAccumulation/tree/master/learn-for-go/grpc-example)
- 上面这个项目的[掘金链接](https://juejin.cn/post/6844904106314694669)
- 另外一个gRPC[例子](https://blog.51cto.com/u_15289640/2966820)

### Structure

### Generate *.pb from *.proto

```bash
protoc --go_out=rpc --go_opt=paths=import \
    --go-grpc_out=rpc --go-grpc_opt=paths=import \
    rpc/proto/*.proto
```

> 如果是希望在同一个目录下面通过*.proto生成pb文件，可以使用:
> ```bash
> protoc --go_out=. --go_opt=paths=source_relative \
>    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
>    rpc/proto/*.proto
> ```

### FAQ

- Q: Timestamp <=> Go Time
  A: [link](https://blog.csdn.net/qq_32828933/article/details/105772122)


### 一些测试路由

#### `/rc/remote-exec`

```bash
curl -d '{"ipv4":"127.0.0.1", "cmd":"node -v"}' -H "Content-Type: application/json" -X POST http://localhost:8091/rc/remote-exec
```

### Example: Normal GRPC

#### Client

```go
import (
	"context"
	"fmt"
	"log"
	StellarTypes "stellar/types"

	pb "stellar/rpc/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
func RegisterGrpcClient(conf *StellarTypes.StellarConfig) {
	conn, err := grpc.Dial(fmt.Sprintf(":%d", 8092), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	c := pb.NewTransferClient(conn)
	r, err := c.TransferSignal(context.Background(), &pb.StellarRequest{
		Type:      "test_request",
		RequestId: "1234",
		Payload:   "ls -la",
	})
	if err != nil {
		panic(err)
	}

	log.Printf("data: %s", r.GetMessage())
}
```

#### Server

```go
import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	StellarTypes "stellar/types"

	pb "stellar/rpc/pb"

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTransferServer
}

func RegisterGrpcServer(conf *StellarTypes.StellarConfig) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterTransferServer(s, &server{})
	startErr := s.Serve(listener)
	if startErr != nil {
		panic(startErr)
	}
}

func (s *server) TransferSignal(ctx context.Context, in *pb.StellarRequest) (*pb.StellarSignal, error) {
	cmd := exec.Command("/bin/bash", "-c", in.GetPayload())
	// stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()

	if err := cmd.Start(); err != nil {
		fmt.Println("Execute failed when Start:" + err.Error())
		panic(1)
	}

	out_bytes, _ := io.ReadAll(stdout)
	stdout.Close()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Execute failed when Wait:" + err.Error())
		panic(err)
	}

	return &pb.StellarSignal{
		Type:      "single_signal",
		Message:   string(out_bytes),
		RequestId: "1235",
		Timestamp: "12345567",
	}, nil
}
```


### Example: Stream GRPC

```go
// full example
package StellarExamples

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	StellarServiceStargate "stellar/service/stargate"
	StellarTypes "stellar/types"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	StellarRPCSignalPb "stellar/rpc/pb"
)

func registerStellar(conf *StellarTypes.StellarConfig) {
	go func() {
		StellarServiceStargate.PushDynamicStatus()
		StellarServiceStargate.PushStaticStatus(conf)
	}()

	// StellarServiceWeb.RegisterStellarWebService(conf)

	conn, err := grpc.Dial(":8089", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcClient := StellarRPCSignalPb.NewTransferClient(conn)

	req := &StellarRPCSignalPb.StellarRequest{}
	// err = jsonpb.UnmarshalString("", req)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(req)

	stream, err := grpcClient.TransferStreamSignal(context.Background(), req)
	if err != nil {
		panic(err)
	}

	for {
		//4.处理服务端发送过来的流信息
		resp, err := stream.Recv()
		if err == io.EOF { //流是否结束
			break
		}
		if err != nil {
			log.Fatalf("clientgetstreamerr:%v", err)
		}
		log.Printf("getfromstreamserver,message:%v,request_id:%v,timestamp:%v", resp.GetMessage(), resp.GetRequestId(), resp.GetTimestamp())
	}
}

type TransferService struct {
	StellarRPCSignalPb.UnimplementedTransferServer
}

func (s *TransferService) TransferStreamSignal(req *StellarRPCSignalPb.StellarRequest, srv StellarRPCSignalPb.Transfer_TransferStreamSignalServer) error {
	for i := 0; i < 5; i++ {
		err := srv.Send(&StellarRPCSignalPb.StellarSignal{
			Type:      "dynamic_status",
			RequestId: fmt.Sprint(i),
			Message:   "hello",
			Timestamp: "not do",
		})
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
func (s *TransferService) TransferSignal(ctx context.Context, req *StellarRPCSignalPb.StellarRequest) (*StellarRPCSignalPb.StellarSignal, error) {
	return &StellarRPCSignalPb.StellarSignal{
		Type:      "nil",
		RequestId: "0",
		Message:   "nil",
		Timestamp: "nil",
	}, nil
}
func registerPlanet(conf *StellarTypes.StellarConfig) {
	listener, err := net.Listen("tcp", ":8089")
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	StellarRPCSignalPb.RegisterTransferServer(grpcServer, &TransferService{})

	err = grpcServer.Serve(listener)
	if err != nil {
		panic(err)
	}
}
```