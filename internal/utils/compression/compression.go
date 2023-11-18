package compression

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/klauspost/compress/zstd"
)

const (
	// GzipType helps identify which compression/decompression to use.
	GzipType = "gzip"

	// ZstdType helps identify which compression/decompression to use.
	ZstdType = "zstd"
)

// CompressWithZstd uses an external dependency for Zstd to compress data and places data in the supplied buffer.
func compressWithZstd(data []byte, buffer *bytes.Buffer) error {

	zstdWriter, err := zstd.NewWriter(buffer)
	if err != nil {
		return err
	}

	_, err = io.Copy(zstdWriter, bytes.NewReader(data))
	if err != nil {

		closeErr := zstdWriter.Close()
		if closeErr != nil {
			return closeErr
		}

		return err
	}

	return zstdWriter.Close()
}

// DecompressWithZstd uses an external dependency for Zstd to decompress data and replaces the supplied buffer with a new buffer with data in it.
func decompressWithZstd(buffer *bytes.Buffer) error {

	zstdReader, err := zstd.NewReader(buffer)
	if err != nil {
		return err
	}
	defer zstdReader.Close()

	data, err := io.ReadAll(zstdReader)
	if err != nil {
		return err
	}

	*buffer = *bytes.NewBuffer(data)

	return nil
}

// CompressWithGzip uses the standard Gzip Writer to compress data and places data in the supplied buffer.
func compressWithGzip(data []byte, buffer *bytes.Buffer) error {

	gzipWriter := gzip.NewWriter(buffer)

	_, err := gzipWriter.Write(data)
	if err != nil {
		return err
	}

	if err := gzipWriter.Close(); err != nil {
		return err
	}

	return nil
}

// DecompressWithGzip uses the standard Gzip Reader to decompress data and replaces the supplied buffer with a new buffer with data in it.
func decompressWithGzip(buffer *bytes.Buffer) error {

	gzipReader, err := gzip.NewReader(buffer)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(gzipReader)
	if err != nil {
		return err
	}

	if err := gzipReader.Close(); err != nil {
		return err
	}

	*buffer = *bytes.NewBuffer(data)

	return nil
}

func Compress(compressionType string, data []byte, buffer *bytes.Buffer) error {

	switch compressionType {
	case ZstdType:
		return compressWithZstd(data, buffer)
	case GzipType:
		fallthrough
	default:
		return compressWithGzip(data, buffer)
	}
}

func Decompress(compressionType string, buffer *bytes.Buffer) error {

	switch compressionType {
	case ZstdType:
		return decompressWithZstd(buffer)
	case GzipType:
		fallthrough
	default:
		return decompressWithGzip(buffer)
	}
}
