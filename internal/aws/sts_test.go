package aws

import "testing"

func TestDevBroker(t *testing.T) {
	broker := DevBroker{}
	creds, err := broker.AssumeRoleWithWebIdentity(AssumeRoleInput{RoleARN: "arn:aws:iam::123:role/test", TTLSeconds: 60})
	if err != nil {
		t.Fatalf("assume role: %v", err)
	}
	if creds.AccessKeyID == "" || creds.SecretAccessKey == "" || creds.SessionToken == "" {
		t.Fatalf("expected credentials")
	}
}
