package testutils

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CreateTestModuleZip creates a ZIP archive containing the given files.
// Files map keys are paths within the archive, values are file contents.
func CreateTestModuleZip(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	for path, content := range files {
		writer, err := zipWriter.Create(path)
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", path, err)
		}
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("failed to write zip entry %s: %v", path, err)
		}
	}

	err := zipWriter.Close()
	if err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

// CreateTestTarGz creates a TAR.GZ archive containing the given files.
// Files map keys are paths within the archive, values are file contents.
func CreateTestTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for path, content := range files {
		// Create tar header
		header := &tar.Header{
			Name:    path,
			Mode:    0644,
			Size:    int64(len(content)),
			ModTime: time.Now(),
		}

		err := tarWriter.WriteHeader(header)
		if err != nil {
			t.Fatalf("failed to write tar header for %s: %v", path, err)
		}

		_, err = tarWriter.Write([]byte(content))
		if err != nil {
			t.Fatalf("failed to write tar content for %s: %v", path, err)
		}
	}

	err := tarWriter.Close()
	if err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes()
}

// ExtractTestArchive extracts a ZIP archive to a temporary directory and returns the path.
// The caller is responsible for cleaning up the returned directory.
func ExtractTestArchive(t *testing.T, archive []byte) string {
	t.Helper()

	tempDir := t.TempDir()

	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		// Create directory for file
		dir := filepath.Join(tempDir, filepath.Dir(file.Name))
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}

		// Create file
		targetPath := filepath.Join(tempDir, file.Name)
		targetFile, err := os.Create(targetPath)
		if err != nil {
			t.Fatalf("failed to create file %s: %v", targetPath, err)
		}

		// Open the file from the zip archive
		rc, err := file.Open()
		if err != nil {
			targetFile.Close()
			t.Fatalf("failed to open zip entry %s: %v", file.Name, err)
		}

		_, err = io.Copy(targetFile, rc)
		rc.Close()
		targetFile.Close()

		if err != nil {
			t.Fatalf("failed to write file %s: %v", targetPath, err)
		}
	}

	return tempDir
}

// ExtractTestTarGz extracts a TAR.GZ archive to a temporary directory and returns the path.
// The caller is responsible for cleaning up the returned directory.
func ExtractTestTarGz(t *testing.T, archive []byte) string {
	t.Helper()

	tempDir := t.TempDir()

	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("failed to read tar entry: %v", err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			dirPath := filepath.Join(tempDir, header.Name)
			err = os.MkdirAll(dirPath, 0755)
			if err != nil {
				t.Fatalf("failed to create directory %s: %v", dirPath, err)
			}

		case tar.TypeReg, tar.TypeRegA:
			// Create file
			targetPath := filepath.Join(tempDir, header.Name)

			// Ensure parent directory exists
			dir := filepath.Dir(targetPath)
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				t.Fatalf("failed to create directory %s: %v", dir, err)
			}

			targetFile, err := os.Create(targetPath)
			if err != nil {
				t.Fatalf("failed to create file %s: %v", targetPath, err)
			}

			_, err = io.CopyN(targetFile, tarReader, header.Size)
			if err != nil && err != io.EOF {
				targetFile.Close()
				t.Fatalf("failed to write file %s: %v", targetPath, err)
			}
			targetFile.Close()
		}
	}

	return tempDir
}

// ListArchiveContents returns a list of file paths contained in a ZIP archive.
func ListArchiveContents(t *testing.T, archive []byte) []string {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}

	var contents []string
	for _, file := range zipReader.File {
		if !file.FileInfo().IsDir() {
			contents = append(contents, file.Name)
		}
	}

	return contents
}

// ReadFileFromArchive reads a specific file from a ZIP archive.
func ReadFileFromArchive(t *testing.T, archive []byte, filename string) []byte {
	t.Helper()

	zipReader, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		t.Fatalf("failed to create zip reader: %v", err)
	}

	for _, file := range zipReader.File {
		if file.Name == filename {
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("failed to open zip entry: %v", err)
			}
			content, err := io.ReadAll(rc)
			rc.Close()
			if err != nil {
				t.Fatalf("failed to read file content: %v", err)
			}
			return content
		}
	}

	t.Fatalf("file %s not found in archive", filename)
	return nil
}

// CreateValidMainTF creates a valid Terraform main.tf file content.
func CreateValidMainTF() string {
	return `variable "instance_type" {
  description = "The type of instance to create"
  type        = string
  default     = "t2.micro"
}

variable "ami_id" {
  description = "The AMI ID to use"
  type        = string
}

resource "aws_instance" "example" {
  ami           = var.ami_id
  instance_type = var.instance_type

  tags = {
    Name = "TestInstance"
  }
}

output "instance_id" {
  description = "The ID of the instance"
  value       = aws_instance.example.id
}
`
}

// CreateREADMEContent creates a standard README.md file content.
func CreateREADMEContent(moduleName string) string {
	return fmt.Sprintf("# %s\n\n"+
		"A Terraform module for managing resources.\n\n"+
		"## Usage\n\n"+
		"```hcl\n"+
		"module \"example\" {\n"+
		"  source = \"./path/to/module\"\n\n"+
		"  instance_type = \"t2.micro\"\n"+
		"  ami_id        = \"ami-12345678\"\n"+
		"}\n"+
		"```\n\n"+
		"## Requirements\n\n"+
		"| Name | Version |\n"+
		"|------|--------|\n"+
		"| terraform | >= 0.12 |\n\n"+
		"## Providers\n\n"+
		"| Name | Version |\n"+
		"|------|--------|\n"+
		"| aws | n/a |\n\n"+
		"## Inputs\n\n"+
		"| Name | Description | Type | Default | Required |\n"+
		"|------|-------------|------|---------|----------|\n"+
		"| instance_type | The type of instance to create | `string` | `\"t2.micro\"` | no |\n"+
		"| ami_id | The AMI ID to use | `string` | n/a | yes |\n\n"+
		"## Outputs\n\n"+
		"| Name | Description |\n"+
		"|------|-------------|\n"+
		"| instance_id | The ID of the instance |\n", moduleName)
}

// CreateTerraregMetadata creates a terrareg.json metadata file content.
func CreateTerraregMetadata(metadata map[string]interface{}) string {
	var builder strings.Builder

	builder.WriteString("{\n")

	first := true
	for key, value := range metadata {
		if !first {
			builder.WriteString(",\n")
		}
		first = false

		// Format key with quotes
		builder.WriteString(fmt.Sprintf("  %q: ", key))

		// Format value based on type
		switch v := value.(type) {
		case string:
			builder.WriteString(fmt.Sprintf("%q", v))
		case bool:
			builder.WriteString(fmt.Sprintf("%t", v))
		case []interface{}:
			builder.WriteString("[")
			for i, item := range v {
				if i > 0 {
					builder.WriteString(", ")
				}
				if str, ok := item.(string); ok {
					builder.WriteString(fmt.Sprintf("%q", str))
				} else {
					builder.WriteString(fmt.Sprintf("%v", item))
				}
			}
			builder.WriteString("]")
		default:
			builder.WriteString(fmt.Sprintf("%v", v))
		}
	}

	builder.WriteString("\n}\n")

	return builder.String()
}
