package server

import (
	"context"
	"exc8/skeleton/pb/exc8/pb"
	"net"
	"sync"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

type GRPCService struct {
	pb.UnimplementedOrderServiceServer

	mu     sync.Mutex
	drinks []*pb.Drink
	orders map[int32]int32
}

// Initialize drinks and orders when service is created
func NewGRPCService() *GRPCService {
	return &GRPCService{
		drinks: []*pb.Drink{
			{Id: 1, Name: "Spritzer", Price: 2, Description: "Wine with soda"},
			{Id: 2, Name: "Beer", Price: 3, Description: "Hagenberger Gold"},
			{Id: 3, Name: "Coffee", Price: 0, Description: "Mifare isn't that secure"},
		},
		orders: make(map[int32]int32),
	}
}

func StartGrpcServer() error {
	// Create a new gRPC server.
	srv := grpc.NewServer()
	// Create grpc service using constructor to initialize drinks and orders
	grpcService := NewGRPCService()
	// Register our service implementation with the gRPC server.
	pb.RegisterOrderServiceServer(srv, grpcService)
	// Serve gRPC server on port 4000.
	lis, err := net.Listen("tcp", ":4000")
	if err != nil {
		return err
	}
	return srv.Serve(lis)
}

// todo implement functions

func (s *GRPCService) GetDrinks(ctx context.Context, _ *emptypb.Empty) (*pb.DrinkList, error) {
	return &pb.DrinkList{Drinks: s.drinks}, nil
}

func (s *GRPCService) OrderDrink(ctx context.Context, req *pb.OrderRequest) (*wrapperspb.BoolValue, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := req.Item
	s.orders[item.DrinkId] += item.Quantity

	return wrapperspb.Bool(true), nil
}

func (s *GRPCService) GetOrders(ctx context.Context, _ *emptypb.Empty) (*pb.AllOrders, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []*pb.OrderItem
	for id, qty := range s.orders {
		items = append(items, &pb.OrderItem{DrinkId: id, Quantity: qty})
	}

	return &pb.AllOrders{Orders: items}, nil
}
