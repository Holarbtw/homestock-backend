package main

import (
	"context"
	"fmt"
	"os"
	"time"

	catalogv1 "github.com/Holarbtw/homestock-backend/pkg/pb/catalog/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const (
	smokeProductID = "11111111-1111-4111-8111-111111111111"
	smokeGTIN      = "04601234567893"
	optionalID     = "22222222-2222-4222-8222-222222222222"
	optionalGTIN   = "03017620422003"
	catalogService = "catalog.v1.CatalogService"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run deploy/smoke-catalog.go HOST:PORT")
		os.Exit(2)
	}
	if err := checkCatalog(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("Catalog gRPC passed: health, error codes, normalized lookup, exact quantity and optional fields.")
}

func checkCatalog(address string) error {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	health, err := healthv1.NewHealthClient(conn).Check(ctx, &healthv1.HealthCheckRequest{Service: catalogService})
	if err != nil {
		return fmt.Errorf("catalog health: %w", err)
	}
	if health.GetStatus() != healthv1.HealthCheckResponse_SERVING {
		return fmt.Errorf("catalog health: expected SERVING, got %s", health.GetStatus())
	}

	client := catalogv1.NewCatalogServiceClient(conn)
	for _, test := range []struct {
		gtin string
		code codes.Code
	}{
		{"invalid-gtin", codes.InvalidArgument},
		{"00000000000001", codes.NotFound},
	} {
		_, err := client.GetProduct(ctx, &catalogv1.GetProductRequest{Gtin: test.gtin})
		if status.Code(err) != test.code {
			return fmt.Errorf("GetProduct(%q): expected %s, got %v", test.gtin, test.code, err)
		}
	}

	short, err := client.GetProduct(ctx, &catalogv1.GetProductRequest{Gtin: "4601234567893"})
	if err != nil {
		return fmt.Errorf("13-digit lookup: %w", err)
	}
	canonical, err := client.GetProduct(ctx, &catalogv1.GetProductRequest{Gtin: smokeGTIN})
	if err != nil {
		return fmt.Errorf("14-digit lookup: %w", err)
	}
	p := canonical.GetProduct()
	if p == nil || !proto.Equal(short.GetProduct(), p) {
		return fmt.Errorf("13- and 14-digit lookups must return the same product")
	}
	if p.GetId() != smokeProductID || p.GetGtin() != smokeGTIN || p.GetName() != "Catalog smoke product" {
		return fmt.Errorf("catalog seed identity did not persist: %v", p)
	}
	if p.GetBrand() != "Smoke brand" || p.GetCategory() != "Smoke category" || p.GetImageUrl() != "https://example.com/catalog-smoke.jpg" {
		return fmt.Errorf("catalog product fields did not map correctly: %v", p)
	}
	if p.GetPackageSize() == nil || p.GetPackageSize().GetQuantity() != "12.345" || p.GetPackageSize().GetUnit() != catalogv1.Unit_UNIT_MILLILITER {
		return fmt.Errorf("package quantity or unit changed: %v", p.GetPackageSize())
	}
	if err := checkProductDates(p); err != nil {
		return err
	}

	optional, err := client.GetProduct(ctx, &catalogv1.GetProductRequest{Gtin: "3017620422003"})
	if err != nil {
		return fmt.Errorf("optional product lookup: %w", err)
	}
	p = optional.GetProduct()
	if p == nil || p.GetId() != optionalID || p.GetGtin() != optionalGTIN || p.GetName() != "Catalog smoke optional product" {
		return fmt.Errorf("optional catalog seed identity did not persist: %v", p)
	}
	if p.Brand != nil || p.Category != nil || p.ImageUrl != nil || p.PackageSize != nil {
		return fmt.Errorf("absent optional fields must remain absent: %v", p)
	}
	return checkProductDates(p)
}

func checkProductDates(product *catalogv1.Product) error {
	if product.CreatedAt == nil || product.UpdatedAt == nil {
		return fmt.Errorf("catalog product timestamps are missing")
	}
	if err := product.CreatedAt.CheckValid(); err != nil {
		return fmt.Errorf("invalid created_at: %w", err)
	}
	if err := product.UpdatedAt.CheckValid(); err != nil {
		return fmt.Errorf("invalid updated_at: %w", err)
	}
	return nil
}
