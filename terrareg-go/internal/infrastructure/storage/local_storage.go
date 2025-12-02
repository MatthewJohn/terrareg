package storage

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// LocalStorage implements the StorageService interface for the local filesystem.
type LocalStorage struct{}

// NewLocalStorage creates a new LocalStorage.
func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

// CopyDir recursively copies a directory from source to destination.
func (s *LocalStorage) CopyDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := s.CopyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Prevent symlink traversal
			if entry.Type()&os.ModeSymlink != 0 {
				continue
			}
			if err := copyFile(srcPath, destPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// MkdirTemp creates a new temporary directory.
func (s *LocalStorage) MkdirTemp(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// RemoveAll removes a path and any children it contains.
func (s *LocalStorage) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Stat returns a FileInfo describing the named file.
func (s *LocalStorage) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

// MkdirAll creates a directory path.
func (s *LocalStorage) MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

// ReadFile reads the file named by filename and returns the contents.
func (s *LocalStorage) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// ReadDir reads the directory named by dirname and returns a list of directory entries.
func (s *LocalStorage) ReadDir(dirname string) ([]fs.DirEntry, error) {
	return os.ReadDir(dirname)
}

// ExtractArchive extracts a ZIP archive from src to dest.
func (s *LocalStorage) ExtractArchive(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Create destination directory
	if err := s.MkdirAll(dest, 0755); err != nil {
		return err
	}

	// Extract each file
	for _, f := range r.File {
		// Prevent path traversal attacks
		if strings.Contains(f.Name, "..") {
			continue
		}

		filePath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			s.MkdirAll(filePath, 0755)
			continue
		}

		// Create parent directories
		if err := s.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		// Create the file
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // Changed to use os.OpenFile with default permissions
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
