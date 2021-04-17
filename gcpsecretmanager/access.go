package gcpsecretmanager

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type SecretManager struct {
	ProjectID string `json:"project_id,omitempty"`
	SecretID  string `json:"secret_id,omitempty"`
	Version   string `json:"version,omitempty"`
}

func (secret *SecretManager) Access() (data []byte, err error) {
	name, err := secret.getVersionName()
	if err != nil {
		return
	}
	data, err = getSecretData(name)
	if err != nil {
		return
	}
	return
}

func (secret *SecretManager) getVersionName() (name string, err error) {
	if secret.ProjectID == "" {
		err = fmt.Errorf("SecretManager.ProjectID is empty.")
		return
	}
	if secret.SecretID == "" {
		err = fmt.Errorf("SecretManager.SecretID is empty.")
		return
	}
	if secret.Version == "" {
		err = fmt.Errorf("SecretManager.Version is empty.")
		return
	}
	name = "projects/" + secret.ProjectID + "/secrets/" + secret.SecretID + "/versions/" + secret.Version
	return
}

func getSecretData(name string) (data []byte, err error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return
	}
	data = result.Payload.Data

	err = client.Close()
	if err != nil {
		return
	}
	return
}
