package service

import (
	"context"
	"fmt"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/mmcloughlin/cb/app/db"
)

// DB opens a database connection to the Cloud SQL instance.
func DB(ctx context.Context) (*db.DB, error) {
	params, err := envs(
		"SQL_CONNECTION_NAME",
		"SQL_DATABASE",
		"SQL_USER",
		"SQL_PASSWORD_SECRET_NAME",
	)
	if err != nil {
		return nil, err
	}

	password, err := secret(ctx, params["SQL_PASSWORD_SECRET_NAME"])
	if err != nil {
		return nil, err
	}

	sock := "/cloudsql/" + params["SQL_CONNECTION_NAME"]
	conn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		sock, params["SQL_DATABASE"], params["SQL_USER"], password)

	return db.Open(conn)
}

func secret(ctx context.Context, name string) (string, error) {
	// Secrets client.
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("create secretmanager client: %w", err)
	}

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("access secret version: %w", err)
	}

	return string(result.Payload.Data), nil
}

// envs looks up multiple environment variables.
func envs(names ...string) (map[string]string, error) {
	values := map[string]string{}
	for _, name := range names {
		v, err := env(name)
		if err != nil {
			return nil, err
		}
		values[name] = v
	}
	return values, nil
}

// env looks up the given evironment variable. Adds the common project prefix.
func env(name string) (string, error) {
	name = "CB_" + name
	v, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("environment variable %s not defined", name)
	}
	return v, nil
}
