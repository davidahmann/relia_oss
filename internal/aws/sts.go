package aws

import "time"

type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	ExpiresAt       time.Time
}

type AssumeRoleInput struct {
	RoleARN    string
	Region     string
	TTLSeconds int
	Subject    string
}

type CredentialBroker interface {
	AssumeRoleWithWebIdentity(input AssumeRoleInput) (Credentials, error)
}

// DevBroker returns placeholder credentials for local development.
type DevBroker struct{}

func (d DevBroker) AssumeRoleWithWebIdentity(input AssumeRoleInput) (Credentials, error) {
	expires := time.Now().UTC().Add(time.Duration(input.TTLSeconds) * time.Second)
	return Credentials{
		AccessKeyID:     "DEV_ACCESS_KEY",
		SecretAccessKey: "DEV_SECRET_KEY",
		SessionToken:    "DEV_SESSION_TOKEN",
		ExpiresAt:       expires,
	}, nil
}
