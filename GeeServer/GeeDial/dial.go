package GeeDial

import (
	"GeeServer/model"
	pb "GeeServer/serverpb"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

func Dial(ip string, content model.Request) (res *pb.Response, err error) {
	fmt.Println("will dial ", ip)
	conn, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		return &pb.Response{}, err
	}
	defer conn.Close()

	c := pb.NewGeeServerClient(conn)

	req := &pb.Request{

		Group: content.Group,
		Key:   content.Key,
		Type:  content.Type,
	}
	//fmt.Printf("Sending request: %+v\n", req) // 添加日志
	res, geterr := c.Get(context.Background(), req)
	if geterr != nil {
		//fmt.Println("dial err: ", geterr)
		return &pb.Response{Err: geterr.Error()}, geterr
	}
	return res, nil
}
