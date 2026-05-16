package storage

type StorageProviderName string

const (
	// ProviderS3 is the provider name for S3-backed storage.
	ProviderS3 = "s3"

	// ProviderMinIO is the provider name for MinIO-backed storage.
	ProviderMinIO = "minio"

	// ProviderLocal is the provider name for local-filesystem-backed storage.
	ProviderLocal = "local"
)
