package server

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	h "net/http"
	v1 "testk/api/helloworld/v1"
	"testk/internal/conf"
	"testk/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	v1.RegisterGreeterHTTPServer(srv, greeter)

	srv.HandleFunc("/books", func(writer h.ResponseWriter, request *h.Request) {
		log.Info("books:", request.Method, request.RequestURI)
		conn, err := grpc.NewClient(c.Grpc.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("could not connect to order service: %v", err)
		}
		defer conn.Close()

		gc := v1.NewGreeterClient(conn)
		stream, err := gc.StreamBooks(context.Background(), &emptypb.Empty{})
		if err != nil {
			h.Error(writer, "error grpc stream", h.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := writer.(h.Flusher)
		if !ok {
			h.Error(writer, "Streaming unsupported!", h.StatusInternalServerError)
			return
		}

		encoder := json.NewEncoder(writer)

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				h.Error(writer, "Failed to receive data from stream: "+err.Error(), h.StatusInternalServerError)
				return
			}

			if err := encoder.Encode(in); err != nil {
				h.Error(writer, "Failed to encode lookup subject response: "+err.Error(), h.StatusInternalServerError)
				return
			}
			flusher.Flush()
		}
	})

	return srv
}
