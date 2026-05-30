package storage

import (
	"context"
	"fmt"
	"time"
)

type Adapter struct {
	providers   map[StorageProviderName]StorageProvider
	defaultName StorageProviderName
}

func NewAdapter() *Adapter {
	return &Adapter{
		providers: make(map[StorageProviderName]StorageProvider),
	}
}

func (a *Adapter) Register(provider StorageProvider) {
	if a.providers == nil {
		a.providers = make(map[StorageProviderName]StorageProvider)
	}
	a.providers[provider.Name()] = provider
	if a.defaultName == "" {
		a.defaultName = provider.Name()
	}
}

func (a *Adapter) SetDefault(name StorageProviderName) error {
	if _, ok := a.providers[name]; !ok {
		return fmt.Errorf("storage: provider %q is not registered", name)
	}
	a.defaultName = name
	return nil
}

func (a *Adapter) HandleUpload(ctx context.Context, req *UploadFileRequest) (*UploadFileResponse, error) {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return nil, err
	}
	return provider.HandleUpload(ctx, req)
}

func (a *Adapter) HandleDelete(ctx context.Context, req *DeleteFileRequest) error {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return err
	}
	return provider.HandleDelete(ctx, req)
}

func (a *Adapter) GetPreviewURL(ctx context.Context, req *GetPreviewURLRequest, duration time.Duration) (string, error) {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return "", err
	}
	return provider.GetPreviewURL(ctx, req, duration)
}

func (a *Adapter) CreatePresignedUrl(ctx context.Context, req *CreatePresignedUrlRequest) (string, error) {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return "", err
	}
	return provider.CreatePresignedUrl(ctx, req)
}

func (a *Adapter) ValidateFileType(ctx context.Context, req *ValidateFileTypeRequest) error {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return err
	}
	return provider.ValidateFileType(ctx, req)
}

func (a *Adapter) Download(ctx context.Context, req *DownloadFileRequest) ([]byte, error) {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return nil, err
	}
	return provider.Download(ctx, req)
}

func (a *Adapter) NormalizeKey(ctx context.Context, req *NormalizeKeyRequest) (string, error) {
	provider, err := a.Get(a.defaultName)
	if err != nil {
		return "", err
	}
	return provider.NormalizeKey(ctx, req)
}

func (a *Adapter) Get(name StorageProviderName) (StorageProvider, error) {
	if provider, ok := a.providers[name]; ok {
		return provider, nil
	}
	return nil, fmt.Errorf("storage: provider %q is not registered", name)
}
