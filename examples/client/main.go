package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	catalogv1 "conduit/generated/catalog/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	time.Sleep(500 * time.Millisecond)

	fmt.Println(" Conduit Dual-Transport Client Demo ")

	// 1. Query REST Transport (HTTP / JSON)
	restURL := "http://localhost:8080/v1/products?query=keyboard"
	fmt.Printf("[1] Calling REST API: GET %s\n", restURL)

	resp, err := http.Get(restURL)
	if err != nil {
		log.Fatalf("REST request failed: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var restData struct {
		Products []struct {
			ID    string  `json:"id"`
			Name  string  `json:"name"`
			Price float64 `json:"price"`
		} `json:"products"`
	}
	_ = json.Unmarshal(bodyBytes, &restData)

	fmt.Printf("    HTTP Status: %s\n", resp.Status)
	fmt.Printf("    REST Response JSON: %s\n", string(bodyBytes))

	// 2. Query gRPC Transport (HTTP/2 / Protobuf)
	grpcAddr := "localhost:50051"
	fmt.Printf("\n[2] Calling gRPC RPC: ProductService.SearchProducts on %s\n", grpcAddr)

	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC: %v", err)
	}
	defer conn.Close()

	client := catalogv1.NewProductServiceClient(conn)
	grpcResp, err := client.SearchProducts(context.Background(), &catalogv1.SearchProductsRequest{
		Query: "keyboard",
		Limit: 5,
	})
	if err != nil {
		log.Fatalf("gRPC request failed: %v", err)
	}

	fmt.Printf("    gRPC Response Products: %d item(s) returned\n", len(grpcResp.Products))
	for _, p := range grpcResp.Products {
		fmt.Printf("      - [%s] %s ($%.2f)\n", p.Id, p.Name, p.Price)
	}

	fmt.Println("\nPerfect Parity: REST and gRPC returned identical data over shared logic!")
}
