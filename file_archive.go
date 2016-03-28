package logrus_rollingfile_hook

import (
	"os"
	"compress/gzip"
	"io"
)

const (
	GzipSuffix = ".gz"
)

type ArchiveFunc func(fileName string) error

// Archiver map used for finding an archive function from a given suffix.
var Archivers = map[string]ArchiveFunc{
	GzipSuffix: gzipArchiveAndDelete,
}

// Gzip file.
func gzipArchive(fileName string) error {
	gzFileName := fileName + GzipSuffix

	// Create .gz file
	gzFile, err := os.OpenFile(gzFileName, os.O_APPEND | os.O_CREATE | os.O_WRONLY, 0664)

	if err != nil {
		return err
	}

	defer gzFile.Close()

	// Create gzip writer
	writer := gzip.NewWriter(gzFile)

	// Open original file for reading
	oldFile, err := os.Open(fileName)

	if err != nil {
		return err
	}

	defer oldFile.Close()

	// Read from original file and write to .gz file
	_, err = io.Copy(writer, oldFile)

	if err != nil {
		return err
	}

	writer.Flush()
	writer.Close()

	return nil
}

// Gzip file and delete.
func gzipArchiveAndDelete(fileName string) error {
	// Gzip file
	err := gzipArchive(fileName)

	if err != nil {
		return err
	}

	// Delete file
	err = os.Remove(fileName)

	if err != nil {
		return err
	}

	return nil
}
