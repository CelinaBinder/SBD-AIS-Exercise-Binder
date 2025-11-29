package client

import (
	"context"
	"exc8/skeleton/pb/exc8/pb"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GrpcClient struct {
	client pb.OrderServiceClient
}

func NewGrpcClient() (*GrpcClient, error) {
	conn, err := grpc.Dial(
		":4000",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	client := pb.NewOrderServiceClient(conn)
	return &GrpcClient{client: client}, nil
}

func (c *GrpcClient) Run() error {
	ctx := context.Background()

	// ---------------- 1. List drinks ----------------
	fmt.Println("Requesting drinks 🍹🍺☕")
	drinksResp, err := c.client.GetDrinks(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	fmt.Println("Available drinks:")
	for _, d := range drinksResp.Drinks {
		fmt.Printf("\t> id:%d  name:\"%s\"  price:%d  description:\"%s\"\n",
			d.Id, d.Name, d.Price, d.Description)
	}

	// Helper function to order drinks
	order := func(drinkID, qty int32) error {
		fmt.Printf("\t> Ordering: %d x %s\n", qty, drinksResp.Drinks[drinkID-1].Name)
		_, err := c.client.OrderDrink(ctx, &pb.OrderRequest{
			Item: &pb.OrderItem{DrinkId: drinkID, Quantity: qty},
		})
		return err
	}

	// ---------------- 2. Order a few drinks ----------------
	fmt.Println("Ordering drinks 👨‍🍳⏱️🍻🍻")
	if err := order(1, 2); err != nil {
		return err
	}
	if err := order(2, 2); err != nil {
		return err
	}
	if err := order(3, 2); err != nil {
		return err
	}

	// ---------------- 3. Order more drinks ----------------
	fmt.Println("Ordering another round of drinks 👨‍🍳⏱️🍻🍻")
	if err := order(1, 6); err != nil {
		return err
	}
	if err := order(2, 6); err != nil {
		return err
	}
	if err := order(3, 6); err != nil {
		return err
	}

	// ---------------- 4. Get order total ----------------
	fmt.Println("Getting the bill")
	allOrders, err := c.client.GetOrders(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}
	for _, o := range allOrders.Orders {
		name := drinksResp.Drinks[o.DrinkId-1].Name
		fmt.Printf("\t> Total: %d x %s\n", o.Quantity, name)
	}

	fmt.Println("Orders complete!")
	return nil

}
