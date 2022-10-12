package vtsclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/veraison/services/config"
	"github.com/veraison/services/proto"
	"github.com/veraison/services/vts/trustedservices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	ErrNoClient = errors.New("there is no active gRPC VTS client")
)

type GRPC struct {
	ServerAddress string
	Connection    *grpc.ClientConn
}

// NewGRPC instantiate a new gRPC store client with the supplied configuration
func NewGRPC() *GRPC {
	return &GRPC{}
}

func (o *GRPC) Init(v *viper.Viper) error {
	cfg := trustedservices.NewGRPCConfig()

	loader := config.NewLoader(cfg)
	if err := loader.LoadFromViper(v); err != nil {
		return err
	}

	o.ServerAddress = cfg.ServerAddress

	return nil
}

func (o *GRPC) AddSwComponents(
	ctx context.Context,
	in *proto.AddSwComponentsRequest,
	opts ...grpc.CallOption,
) (*proto.AddSwComponentsResponse, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, fmt.Errorf("failed AddSwComponents: %w", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.AddSwComponents(ctx, in, opts...)
}

func (o *GRPC) AddTrustAnchor(
	ctx context.Context,
	in *proto.AddTrustAnchorRequest,
	opts ...grpc.CallOption,
) (*proto.AddTrustAnchorResponse, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, fmt.Errorf("failed AddTrustAnchor: %w", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.AddTrustAnchor(ctx, in, opts...)
}

func (o *GRPC) GetAttestation(
	ctx context.Context, in *proto.AttestationToken, opts ...grpc.CallOption,
) (*proto.AppraisalContext, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, fmt.Errorf("failed GetAttestation: %w", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetAttestation(ctx, in, opts...)
}

func (o *GRPC) GetSupportedVerificationMediaTypes(
	ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption,
) (*proto.MediaTypeList, error) {
	if err := o.EnsureConnection(); err != nil {
		return nil, fmt.Errorf("failed GetSupportedVerificationMediaTypes: %w", err)
	}

	c := o.GetProvisionerClient()
	if c == nil {
		return nil, ErrNoClient
	}

	return c.GetSupportedVerificationMediaTypes(ctx, in, opts...)
}

func (o *GRPC) GetProvisionerClient() proto.VTSClient {
	if o.Connection == nil {
		return nil
	}

	return proto.NewVTSClient(o.Connection)
}

func (o *GRPC) EnsureConnection() error {
	if o.Connection != nil {
		return nil
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, o.ServerAddress, opts...)
	if err != nil {
		return fmt.Errorf("connection to gRPC VTS server %s failed: %w", o.ServerAddress, err)
	}

	o.Connection = conn

	return nil
}
