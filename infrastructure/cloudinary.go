package infrastructure

import (
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

func NewCloudinary() (*cloudinary.Cloudinary, error) {
	if err := validateCloudinaryEnv(); err != nil {
		return nil, err
	}

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}

	return cld, nil
}

func validateCloudinaryEnv() error {
	required := map[string]string{
		"CLOUDINARY_CLOUD_NAME": os.Getenv("CLOUDINARY_CLOUD_NAME"),
		"CLOUDINARY_API_KEY":    os.Getenv("CLOUDINARY_API_KEY"),
		"CLOUDINARY_API_SECRET": os.Getenv("CLOUDINARY_API_SECRET"),
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("environment variable %s is not set", key)
		}
	}
	return nil
}