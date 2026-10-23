package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	catalogv1 "conduit/generated/catalog/v1"

	"google.golang.org/grpc"
)

var catalog = []*catalogv1.Product{
	{Id: "p1", Name: "Mechanical Keyboard", Price: 129.99},
	{Id: "p2", Name: "Wireless Gaming Mouse", Price: 79.99},
	{Id: "p3", Name: "UltraWide 4K Monitor", Price: 499.99},
	{Id: "p4", Name: "Noise-Cancelling Headphones", Price: 199.99},
	{Id: "p5", Name: "Ergonomic Desk Chair", Price: 349.99},
}

func searchProducts(query string, limit int32) []*catalogv1.Product {
	query = strings.ToLower(strings.TrimSpace(query))
	var results []*catalogv1.Product

	for _, item := range catalog {
		if query == "" || strings.Contains(strings.ToLower(item.Name), query) {
			results = append(results, item)
		}
		if limit > 0 && int32(len(results)) >= limit {
			break
		}
	}
	return results
}

// --- Transport 1: gRPC Implementation ---

type grpcServer struct {
	catalogv1.UnimplementedProductServiceServer
}

func (s *grpcServer) SearchProducts(ctx context.Context, req *catalogv1.SearchProductsRequest) (*catalogv1.SearchProductsResponse, error) {
	log.Printf("[gRPC] Received SearchProducts: query=%q, limit=%d", req.Query, req.Limit)
	items := searchProducts(req.Query, req.Limit)
	return &catalogv1.SearchProductsResponse{Products: items}, nil
}

func startGRPCServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen for gRPC on port %d: %v", port, err)
	}

	s := grpc.NewServer()
	catalogv1.RegisterProductServiceServer(s, &grpcServer{})

	log.Printf("gRPC server listening on port %d...", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}

// --- Transport 2: REST / HTTP Implementation ---

type restResponse struct {
	Products []*catalogv1.Product `json:"products"`
}

func handleRESTSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")

	var limit int32
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = int32(val)
		}
	}

	log.Printf("[REST] Received GET /v1/products: query=%q, limit=%d", query, limit)
	items := searchProducts(query, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(restResponse{Products: items}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func startRESTServer(port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/products", handleRESTSearch)

	log.Printf("REST HTTP server listening on port %d ...", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("REST server error: %v", err)
	}
}

func main() {
	grpcPort := 50051
	restPort := 8080

	fmt.Println(" Conduit Dual-Transport Server (gRPC + REST)")

	go startGRPCServer(grpcPort)

	startRESTServer(restPort)
}
