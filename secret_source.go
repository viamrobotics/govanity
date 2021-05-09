package govanity

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/erh/egoutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrSecretNotFound = errors.New("secret not found")

type SecretSource interface {
	Get(ctx context.Context, name string) (string, error)
	Type() SecretSourceType
}

type SecretSourceType string

const (
	SecretSourceTypeEnv = "env"
	SecretSourceTypeGCP = "gcp"
)

func NewSecretSource(ctx context.Context, sourceType SecretSourceType) (SecretSource, error) {
	switch sourceType {
	case SecretSourceTypeGCP:
		return NewGCPSecretSource(ctx)
	case "", SecretSourceTypeEnv:
		return &EnvSecretSource{}, nil
	default:
		return nil, fmt.Errorf("unknown secret source type %q", sourceType)
	}
}

type EnvSecretSource struct {
}

func (src *EnvSecretSource) Get(ctx context.Context, name string) (string, error) {
	secret, ok := os.LookupEnv(name)
	if !ok {
		return "", ErrSecretNotFound
	}
	return secret, nil
}

func (src *EnvSecretSource) Type() SecretSourceType {
	return SecretSourceTypeEnv
}

func NewGCPSecretSource(ctx context.Context) (*GCPSecretSource, error) {
	gcpSecrets, err := egoutil.NewGCPSecrets(ctx)
	if err != nil {
		return nil, err
	}
	return &GCPSecretSource{gcpSecrets}, nil
}

type GCPSecretSource struct {
	gcpSecrets *egoutil.GCPSecrets
}

func (src *GCPSecretSource) Get(ctx context.Context, name string) (string, error) {
	secret, err := src.gcpSecrets.GetSecret(ctx, name)
	if err != nil {
		if status.Convert(errors.Unwrap(err)).Code() == codes.NotFound {
			return "", ErrSecretNotFound
		}
		return "", err
	}
	return secret, nil
}

func (src *GCPSecretSource) Type() SecretSourceType {
	return SecretSourceTypeGCP
}
