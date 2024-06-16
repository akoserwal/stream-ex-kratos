package service

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/emptypb"

	v1 "testk/api/helloworld/v1"
	"testk/internal/biz"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello implements helloworld.GreeterServer.
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}

func (s GreeterService) StreamBooks(empy *emptypb.Empty, stream v1.Greeter_StreamBooksServer) error {
	fmt.Println("stream books")
	books := []*v1.Book{
		{Id: "1", Title: "1984", Author: "George Orwell"},
		{Id: "2", Title: "To Kill a Mockingbird", Author: "Harper Lee"},
	}

	for _, book := range books {
		if err := stream.Send(book); err != nil {
			return err
		}
	}
	return nil
}
