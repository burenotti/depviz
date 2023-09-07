package app

import "fmt"

const (
	Npm = "npm"
	Pip = "pip"
)

type Config struct {
	PackageName    string
	PackageManager string
}

func (c *Config) Validate() error {
	if c.PackageName == "" {
		return fmt.Errorf("package name is required")
	}

	if c.PackageManager == "" {
		return fmt.Errorf("package manager is required")
	}

	if c.PackageManager != Npm && c.PackageManager != Pip {
		return fmt.Errorf("package manager is invalid")
	}
	return nil
}
