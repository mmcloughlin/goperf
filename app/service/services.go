// Package service initializes Google Cloud services.
package service

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"

	"github.com/mmcloughlin/cb/app/db"
	"github.com/mmcloughlin/cb/app/gcs"
	"github.com/mmcloughlin/cb/pkg/fs"
)

// Init is an initializaiton function.
type Init func(ctx context.Context, l *zap.Logger) error

// Initialize runs an initializaiton function.
func Initialize(i Init) {
	ctx := context.Background()

	logger, err := Logger()
	if err != nil {
		log.Fatal(err)
	}

	if err := i(ctx, logger); err != nil {
		logger.Error("initialization error", zap.Error(err))
		os.Exit(1)
	}
}

// Logger builds a logger for use in service code.
func Logger() (*zap.Logger, error) {
	return zapdriver.NewProduction()
}

// ResultsFileSystem builds filesystem access to results data files.
func ResultsFileSystem(ctx context.Context) (fs.Interface, error) {
	bucket, err := env("RESULTS_BUCKET")
	if err != nil {
		return nil, err
	}
	return gcs.New(ctx, bucket)
}

// DB opens a database connection to the Cloud SQL instance.
func DB(ctx context.Context, l *zap.Logger) (*db.DB, error) {
	params, err := envs(
		"SQL_IP_ADDRESS",
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

	conn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s",
		params["SQL_IP_ADDRESS"], params["SQL_DATABASE"], params["SQL_USER"], password)

	d, err := db.Open(ctx, conn)
	if err != nil {
		return nil, err
	}
	d.SetLogger(l)
	return d, nil
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
