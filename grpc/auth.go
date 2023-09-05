package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Token struct {
	AppSecret string
}

func (t *Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"app_secret": t.AppSecret}, nil
}

func (t *Token) RequireTransportSecurity() bool {
	return true
}

func InjectGrpcCustomAuth(secret string, host string, port int, usingTls bool, pubKey string, serverName string) (*grpc.ClientConn, error) {
	token := Token{
		AppSecret: secret,
	}
	var (
		conn    *grpc.ClientConn
		grpcErr error
	)
	if usingTls {
		cred, err := credentials.NewClientTLSFromFile(pubKey, serverName)
		if err != nil {
			panic(err)
		}
		conn, grpcErr = grpc.Dial(fmt.Sprintf("%s:%v", host, port), grpc.WithTransportCredentials(cred), grpc.WithPerRPCCredentials(&token))
	} else {
		conn, grpcErr = grpc.Dial(fmt.Sprintf("%s:%v", host, port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithPerRPCCredentials(&token))
	}
	defer conn.Close()
	return conn, grpcErr
}

func VerifyGrpcCustomAuth(ctx context.Context, password string) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	var AppSecret string
	if val, ok := md["app_secret"]; ok {
		AppSecret = val[0]
	}
	if AppSecret != password {
		return false
	}
	return true
}
