package client

import (
	"context"
	"time"

	compassv1beta1 "github.com/goto/compass/proto/gotocompany/compass/v1beta1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Config struct {
	Host                   string `mapstructure:"host" default:"localhost:8081"`
	ServerHeaderKeyEmail   string `yaml:"serverheaderkey_email" mapstructure:"serverheaderkey_email" default:"Compass-User-Email"`
	ServerHeaderValueEmail string `yaml:"serverheadervalue_email" mapstructure:"serverheadervalue_email" default:"compass@gotocompany.com"`
}

func Create(ctx context.Context, cfg Config) (compassv1beta1.CompassServiceClient, func(), error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(ctx, time.Second*2)
	conn, err := createConnection(dialTimeoutCtx, cfg)
	if err != nil {
		dialCancel()
		return nil, nil, err
	}

	cancel := func() {
		dialCancel()
		conn.Close()
	}

	client := compassv1beta1.NewCompassServiceClient(conn)
	return client, cancel, nil
}

func SetMetadata(ctx context.Context, cfg Config) context.Context {
	md := metadata.New(map[string]string{cfg.ServerHeaderKeyEmail: cfg.ServerHeaderValueEmail})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

func createConnection(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	return grpc.DialContext(
		ctx,
		cfg.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)
}
