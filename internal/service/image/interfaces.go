package image

import (
	"apubot/internal/domain"
	"context"
)

type ImageService interface {
	GetRandomFile(ctx context.Context) (file domain.File)
	UpdateFile(ctx context.Context, file domain.File) error
}

type ImageRepository interface {
	GetAll(ctx context.Context) (map[string]string, error)
	SaveImage(ctx context.Context, file domain.File) error
}
