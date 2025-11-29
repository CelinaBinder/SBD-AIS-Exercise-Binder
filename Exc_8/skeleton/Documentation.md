# 📘 Exercise 8 --- gRPC Client and Server

## 1. Introduction

The goal of this exercise is to build a simple gRPC-based drink ordering
system. The system consists of:

-   A **gRPC server** that exposes an `OrderService` with three RPC
    endpoints:
    -   `GetDrinks` --- return all available drinks
    -   `OrderDrink` --- store an order in memory
    -   `GetOrders` --- return aggregated orders
-   A **gRPC client** that:
    -   Lists available drinks
    -   Places multiple rounds of drink orders
    -   Retrieves the final bill
-   A **Protobuf schema (`orders.proto`)** used to generate Go
    server/client stubs.

The server stores all data **in memory**, and no database is required.

------------------------------------------------------------------------

## 2. Installing Protobuf Tools

### Install protoc plugin for Go gRPC:

    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

### Add Go binaries to PATH:

    $env:Path += ";$env:USERPROFILE\go\bin"

### Verify installation:

    protoc-gen-go --version

### Generate Protobuf files:

    ./generate_pb_simple.ps1

------------------------------------------------------------------------

## 3. Protobuf Definition (`orders.proto`)

The `orders.proto` file defines all required messages and RPC methods.

### Messages implemented:

-   `Drink`
-   `DrinkList`
-   `OrderItem`
-   `OrderRequest`
-   `AllOrders`

### Service definition:

``` proto
service OrderService {
  rpc OrderDrink (OrderRequest) returns (google.protobuf.BoolValue);
  rpc GetDrinks (google.protobuf.Empty) returns (DrinkList);
  rpc GetOrders (google.protobuf.Empty) returns (AllOrders);
}
```

### go_package option

    option go_package = "exc8/pb";

------------------------------------------------------------------------

## 4. Generating Go Code from Protobuf

Using the PowerShell script:

    generate_pb_simple.ps1

This runs:

    protoc --go_out=. --go-grpc_out=. orders.proto

Output files:

-   `orders.pb.go`
-   `orders_grpc.pb.go`

These contain:

✔ Data structures\
✔ Client interfaces\
✔ Server interfaces\
✔ RPC registration functions

------------------------------------------------------------------------

## 5. Implementing the gRPC Server

File: `server/grpc_server.go`

### Server struct:

``` go
type GRPCService struct {
    pb.UnimplementedOrderServiceServer
    mu     sync.Mutex
    drinks []*pb.Drink
    orders map[int32]int32
}
```

### Prepopulated drinks:

``` go
drinks: []*pb.Drink{
    {Id: 1, Name: "Spritzer", Price: 2, Description: "Wine with soda"},
    {Id: 2, Name: "Beer", Price: 3, Description: "Hagenberger Gold"},
    {Id: 3, Name: "Coffee", Price: 0, Description: "Mifare isn't that secure"},
},
```

### RPC Implementations

#### GetDrinks

``` go
func (s *GRPCService) GetDrinks(ctx context.Context, _ *emptypb.Empty) (*pb.DrinkList, error) {
    return &pb.DrinkList{Drinks: s.drinks}, nil
}
```

#### OrderDrink

``` go
func (s *GRPCService) OrderDrink(ctx context.Context, req *pb.OrderRequest) (*wrapperspb.BoolValue, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    item := req.Item
    s.orders[item.DrinkId] += item.Quantity

    return wrapperspb.Bool(true), nil
}
```

#### GetOrders

``` go
func (s *GRPCService) GetOrders(ctx context.Context, _ *emptypb.Empty) (*pb.AllOrders, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    var items []*pb.OrderItem
    for id, qty := range s.orders {
        items = append(items, &pb.OrderItem{DrinkId: id, Quantity: qty})
    }

    return &pb.AllOrders{Orders: items}, nil
}
```

------------------------------------------------------------------------

## 6. Implementing the gRPC Client

File: `client/grpc_client.go`

### Create connection:

``` go
conn, err := grpc.Dial(":4000", grpc.WithTransportCredentials(insecure.NewCredentials()))
```


------------------------------------------------------------------------

## 7. Starting Server and Client

File: `main.go`

### Start server:

``` go
go func() {
    if err := server.StartGrpcServer(); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}()
```

### Wait for server:

    time.Sleep(1 * time.Second)

### Start client:

``` go
c, _ := client.NewGrpcClient()
c.Run()
```

------------------------------------------------------------------------

## 8. Final Output

    Requesting drinks 🍹🍺☕
    Available drinks:
        > id:1  name:"Spritzer"  price:2  description:"Wine with soda"
        > id:2  name:"Beer"  price:3  description:"Hagenberger Gold"
        > id:3  name:"Coffee"  price:0  description:"Mifare isn't that secure"
    Ordering drinks 👨‍🍳⏱️🍻🍻
        > Ordering: 2 x Spritzer
        > Ordering: 2 x Beer
        > Ordering: 2 x Coffee
    Ordering another round of drinks 👨‍🍳⏱️🍻🍻
        > Ordering: 6 x Spritzer
        > Ordering: 6 x Beer
        > Ordering: 6 x Coffee
        Getting the bill
        > Total: 8 x Coffee
        > Total: 8 x Spritzer
        > Total: 8 x Beer
    Orders complete!
    Orders complete!

------------------------------------------------------------------------


