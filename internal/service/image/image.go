package image

import (
	"apubot/internal/config"
	"apubot/internal/domain"
	"context"
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type Service struct {
	cfg            *config.Config
	repo           ImageRepository
	availableFiles map[string]string
	mu             sync.RWMutex
}

func New(cfg *config.Config, repo ImageRepository) *Service {
	service := &Service{
		cfg:            cfg,
		repo:           repo,
		availableFiles: make(map[string]string),
	}

	err := service.updateAvailableFiles()
	if err != nil {
		log.Fatalf("can not initialize Image service: %v", err)
	}

	return service
}

func (s *Service) updateAvailableFiles() error {
	var imageFiles map[string]string
	supportedExtensions := []string{".jpg", ".jpeg", ".png", ".gif"}

	imageFiles, err := s.repo.GetAll(context.Background())
	if err != nil {
		return errors.Wrap(err, "can not read data from db")
	}

	filesFs, err := os.ReadDir(s.cfg.ImagesDirPath)
	if err != nil {
		return errors.Wrap(err, "can not read directory")
	}

	for _, fileFs := range filesFs {
		if fileFs.IsDir() {
			continue
		}

		ext := filepath.Ext(fileFs.Name())
		if !slices.Contains(supportedExtensions, ext) {
			continue
		}

		_, ok := imageFiles[fileFs.Name()]
		if !ok {
			imageFiles[fileFs.Name()] = ""
		}
	}

	if len(imageFiles) == 0 {
		return errors.New("no available images in selected directory or db")
	}

	s.availableFiles = imageFiles

	return nil
}

func (s *Service) GetRandomFile(ctx context.Context) (domain.File, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n := rand.Intn(len(s.availableFiles))

	for k, v := range s.availableFiles {
		if n == 0 {
			return domain.File{Name: k, TgID: v}, nil
		}

		n--
	}

	// this should never happen
	return domain.File{}, errors.New("GetRandomFile: no file selected")
}

func (s *Service) UpdateFile(ctx context.Context, file domain.File) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.repo.SaveImage(ctx, file)
	if err != nil {
		return errors.Wrap(err, "can not update image")
	}

	s.availableFiles[file.Name] = file.TgID

	return nil
}
