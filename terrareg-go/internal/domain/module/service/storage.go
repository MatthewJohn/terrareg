package service

import "io/fs"

// StorageService defines the interface for file storage operations.
type StorageService interface {
	// CopyDir recursively copies a directory from source to destination.
	CopyDir(src, dest string) error
	// MkdirTemp creates a new temporary directory.
	MkdirTemp(dir, pattern string) (string, error)
	// RemoveAll removes a path and any children it contains.
	RemoveAll(path string) error
	// Stat returns a FileInfo describing the named file.
	Stat(name string) (fs.FileInfo, error)
	// MkdirAll creates a directory path.
	MkdirAll(path string, perm fs.FileMode) error
	// ReadFile reads the file named by filename and returns the contents.
	ReadFile(filename string) ([]byte, error)
	// ReadDir reads the directory named by dirname and returns a list of directory entries.
	ReadDir(dirname string) ([]fs.DirEntry, error)
	// ExtractArchive extracts a ZIP archive from src to dest.
	ExtractArchive(src, dest string) error
}
