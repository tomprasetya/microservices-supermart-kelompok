package benchmark

import (
	"encoding/json"
	"fmt"
	"testing"

	inventoryv1 "github.com/tomprasetya/microservices-supermart-kelompok/proto/inventory/v1"
	"google.golang.org/protobuf/proto"
)

func TestPayloadComparison(t *testing.T) {
	// Muatan biner Protobuf
	pbItem := &inventoryv1.StockItem{
		ProductId:        "PROD-001",
		WarehouseId:      "WH-BALI-01",
		QuantityOnHand:   150,
		QuantityReserved: 12,
		UnitPrice:        &inventoryv1.Money{CurrencyCode: "IDR", Amount: 75000},
	}
	pbRaw, err := proto.Marshal(pbItem)
	if err != nil {
		t.Fatalf("Gagal marshal protobuf: %v", err)
	}

	// Muatan teks REST JSON
	jsonMap := map[string]interface{}{
		"product_id":        "PROD-001",
		"warehouse_id":      "WH-BALI-01",
		"quantity_on_hand":   150,
		"quantity_reserved": 12,
		"unit_price": map[string]interface{}{
			"currency_code": "IDR",
			"amount":        75000,
		},
	}
	jsonRaw, err := json.Marshal(jsonMap)
	if err != nil {
		t.Fatalf("Gagal marshal json: %v", err)
	}

	efficiency := float64(len(jsonRaw)-len(pbRaw)) / float64(len(jsonRaw)) * 100

	fmt.Printf("\n========================================\n")
	fmt.Printf("HASIL AUDIT UKURAN PAYLOAD (BYTES)\n")
	fmt.Printf("========================================\n")
	fmt.Printf("REST JSON (HTTP/1.1)    : %d bytes\n", len(jsonRaw))
	fmt.Printf("gRPC Protobuf (HTTP/2)  : %d bytes\n", len(pbRaw))
	fmt.Printf("Efisiensi Penghematan   : %.2f%% lebih ringkas\n", efficiency)
	fmt.Printf("========================================\n\n")
}