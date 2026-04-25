package catalog

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/badAkne/order-service/internal/catalog/gen/proto/v1"
)

type CatalogClient struct {
	conn   *grpc.ClientConn
	client pb.CatalogServiceClient
}

func NewCatalogClient(addr string) (*CatalogClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor,
			otelgrpc.UnaryClientInterceptor(), //nolint:staticcheck
		),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor,
			otelgrpc.StreamClientInterceptor(), //nolint:staticcheck
		),
	)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to connec to grpc Client")
		return nil, fmt.Errorf("failed to connect to catalog service: %w", err)
	}

	client := pb.NewCatalogServiceClient(conn)

	log.Info().Msgf("Connected to catalog service at: %s", addr)

	return &CatalogClient{
		conn:   conn,
		client: client,
	}, nil
}

func (c *CatalogClient) Closer() error {
	if c.conn != nil {
		log.Info().Msg("Closing connecction to Catalog service")

		return c.conn.Close()
	}

	return nil
}

type ProductInfo struct {
	ID           string
	GUID         string
	Name         string
	Description  string
	Price        float64
	CategoryGUID string
}

func (c *CatalogClient) GetProducts(ctx context.Context, productsGUIDs []uuid.UUID) (map[string]*ProductInfo, error) {
	guids := make([]string, len(productsGUIDs))

	for i, guid := range productsGUIDs {
		guids[i] = guid.String()
	}

	req := &pb.GetProductsRequest{
		Guid: guids,
	}

	res, err := c.client.GetProducts(ctx, req)
	if err != nil {
		return nil, c.handleError(err)
	}

	products := make(map[string]*ProductInfo)

	for _, product := range res.Products {
		products[product.Guid] = &ProductInfo{
			GUID:         product.Guid,
			Name:         product.Name,
			Description:  product.Description,
			Price:        product.Price,
			CategoryGUID: product.CategoryGUID,
		}
	}

	return products, nil
}

func (c *CatalogClient) CheckProductExists(ctx context.Context, productGUID uuid.UUID) (bool, float64, error) {
	req := &pb.CheckProductExistsRequest{
		Guid: productGUID.String(),
	}

	res, err := c.client.CheckProductExists(ctx, req)
	if err != nil {
		return false, 0, c.handleError(err)
	}

	return res.Exists, res.Price, nil
}

func (c *CatalogClient) handleError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("catalog service error: %w", err)
	}

	switch st.Code() {
	case codes.NotFound:
		return fmt.Errorf("product not found on catalog")
	case codes.Unavailable:
		return fmt.Errorf("catalog service unavailaable")
	case codes.InvalidArgument:
		return fmt.Errorf("invalid request to catalog service: %s", st.Message())
	default:
		return fmt.Errorf("catalog service error [%s]:%s", st.Code(), st.Message())
	}
}
