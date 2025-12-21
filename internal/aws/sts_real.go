package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSBroker struct {
	client stsAssumer
}

const maxInt32 = int(^uint32(0) >> 1)

func NewSTSBroker(region string) (*STSBroker, error) {
	if strings.TrimSpace(region) == "" {
		return nil, fmt.Errorf("missing region")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}
	return &STSBroker{client: sts.NewFromConfig(cfg)}, nil
}

func (b *STSBroker) AssumeRoleWithWebIdentity(input AssumeRoleInput) (Credentials, error) {
	if input.RoleARN == "" {
		return Credentials{}, fmt.Errorf("missing role arn")
	}
	if input.WebIdentityToken == "" {
		return Credentials{}, fmt.Errorf("missing web identity token")
	}
	if input.TTLSeconds <= 0 {
		return Credentials{}, fmt.Errorf("invalid ttl")
	}
	if input.TTLSeconds > maxInt32 {
		return Credentials{}, fmt.Errorf("invalid ttl")
	}

	sessionName := "relia"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := b.client.AssumeRoleWithWebIdentity(ctx, &sts.AssumeRoleWithWebIdentityInput{
		RoleArn:          &input.RoleARN,
		RoleSessionName:  &sessionName,
		WebIdentityToken: &input.WebIdentityToken,
		DurationSeconds:  int32Ptr(int32(input.TTLSeconds)),
	})
	if err != nil {
		return Credentials{}, err
	}
	if out.Credentials == nil {
		return Credentials{}, fmt.Errorf("missing credentials")
	}
	creds := out.Credentials
	expiresAt := time.Now().UTC()
	if creds.Expiration != nil {
		expiresAt = *creds.Expiration
	}

	return Credentials{
		AccessKeyID:     strOrEmpty(creds.AccessKeyId),
		SecretAccessKey: strOrEmpty(creds.SecretAccessKey),
		SessionToken:    strOrEmpty(creds.SessionToken),
		ExpiresAt:       expiresAt,
	}, nil
}

type stsAssumer interface {
	AssumeRoleWithWebIdentity(ctx context.Context, params *sts.AssumeRoleWithWebIdentityInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithWebIdentityOutput, error)
}

func int32Ptr(v int32) *int32 { return &v }
func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
