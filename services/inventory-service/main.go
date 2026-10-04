package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	inventoryv1 "github.com/tomprasetya/microservices-supermart-kelompok/proto/inventory/v1"
	"google.golang.org/grpc"
)

var mockInventory = map[string]*inventoryv1.StockItem{
	"PROD-001": {
		ProductId:        "PROD-001",
		WarehouseId:      "WH-BALI-01",
		QuantityOnHand:   150,
		QuantityReserved: 12,
		UnitPrice: &inventoryv1.Money{
			CurrencyCode: "IDR",
			Amount:       75000,
		},
	},
	"PROD-002": {
		ProductId:        "PROD-002",
		WarehouseId:      "WH-BALI-01",
		QuantityOnHand:   80,
		QuantityReserved: 5,
		UnitPrice: &inventoryv1.Money{
			CurrencyCode: "IDR",
			Amount:       120000,
		},
	},
}

// Handler gRPC
type grpcServer struct {
	inventoryv1.UnimplementedInventoryServiceServer
}

func (s *grpcServer) GetStockItem(ctx context.Context, req *inventoryv1.GetStockItemRequest) (*inventoryv1.GetStockItemResponse, error) {
	item, ok := mockInventory[req.GetProductId()]
	if !ok {
		item = &inventoryv1.StockItem{
			ProductId:   req.GetProductId(),
			WarehouseId: req.GetWarehouseId(),
			UnitPrice:   &inventoryv1.Money{CurrencyCode: "IDR", Amount: 0},
		}
	}
	return &inventoryv1.GetStockItemResponse{Item: item}, nil
}

func (s *grpcServer) BatchGetStockItems(ctx context.Context, req *inventoryv1.BatchGetStockItemsRequest) (*inventoryv1.BatchGetStockItemsResponse, error) {
	var items []*inventoryv1.StockItem
	for _, id := range req.GetProductIds() {
		if it, exists := mockInventory[id]; exists {
			items = append(items, it)
		}
	}
	return &inventoryv1.BatchGetStockItemsResponse{Items: items}, nil
}

// Handler REST
type RestStockResponse struct {
	ProductID        string `json:"product_id"`
	WarehouseID      string `json:"warehouse_id"`
	QuantityOnHand   int32  `json:"quantity_on_hand"`
	QuantityReserved int32  `json:"quantity_reserved"`
	UnitPrice        struct {
		CurrencyCode string `json:"currency_code"`
		Amount       int64  `json:"amount"`
	} `json:"unit_price"`
}

func restStockHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("product_id")
	item, ok := mockInventory[id]

	var resp RestStockResponse
	if ok {
		resp.ProductID = item.ProductId
		resp.WarehouseID = item.WarehouseId
		resp.QuantityOnHand = item.QuantityOnHand
		resp.QuantityReserved = item.QuantityReserved
		resp.UnitPrice.CurrencyCode = item.UnitPrice.CurrencyCode
		resp.UnitPrice.Amount = item.UnitPrice.Amount
	} else {
		resp.ProductID = id
		resp.WarehouseID = "WH-BALI-01"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	// Jalankan gRPC Server pada Port 50051
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			log.Fatalf("Gagal inisialisasi listener gRPC: %v", err)
		}
		s := grpc.NewServer()
		inventoryv1.RegisterInventoryServiceServer(s, &grpcServer{})
		fmt.Println("[gRPC] Inventory Service siap di port :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Gagal serve gRPC: %v", err)
		}
	}()

	// Jalankan REST Server pada Port 8080
	http.HandleFunc("/api/v1/stock", restStockHandler)
	fmt.Println("[REST] Inventory Service siap di port :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatalf("Gagal serve REST: %v", err)
	}
}