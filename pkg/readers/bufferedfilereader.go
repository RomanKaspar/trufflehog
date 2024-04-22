package readers

import (
	"fmt"
	"io"

	"github.com/trufflesecurity/trufflehog/v3/pkg/context"
	bufferedfilewriter "github.com/trufflesecurity/trufflehog/v3/pkg/writers/buffered_file_writer"
)

// bufferedFileReader provides random access read, seek, and close capabilities on top of the BufferedFileWriter.
// It combines the functionality of BufferedFileWriter for buffered writing, bytes.Reader for
// random access reading and seeking, and adds a close method to return the buffer to the pool.
type bufferedFileReader struct {
	bufWriter *bufferedfilewriter.BufferedFileWriter
	reader    io.ReadSeekCloser
}

// NewBufferedFileReader initializes a bufferedFileReader from an io.Reader by using
// the BufferedFileWriter's functionality to read and store data, then setting up a bytes.Reader for random access.
// It returns the initialized bufferedFileReader and any error encountered during the process.
func NewBufferedFileReader(ctx context.Context, r io.Reader) (*bufferedFileReader, error) {
	writer, err := bufferedfilewriter.NewFromReader(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("error creating bufferedFileReader: %w", err)
	}

	// Ensure that the BufferedFileWriter is in read-only mode.
	if err := writer.CloseForWriting(); err != nil {
		return nil, err
	}

	reader, err := writer.ReadCloser()
	if err != nil {
		return nil, err
	}

	rdr, ok := reader.(io.ReadSeekCloser)
	if !ok {
		return nil, fmt.Errorf("reader does not implement io.ReadSeekCloser")
	}

	return &bufferedFileReader{writer, rdr}, nil
}

// Close releases the buffer back to the buffer pool.
// It should be called when the bufferedFileReader is no longer needed.
// Note that closing the bufferedFileReader does not affect the underlying bytes.Reader,
// which can still be used for reading, seeking, and reading at specific positions.
// Close is a no-op for the bytes.Reader.
func (b *bufferedFileReader) Close() error {
	return b.reader.Close()
}

// Read reads up to len(p) bytes into p from the underlying bytes.Reader.
// It returns the number of bytes read and any error encountered.
// If the bytes.Reader reaches the end of the available data, Read returns 0, io.EOF.
// It implements the io.Reader interface.
func (b *bufferedFileReader) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

// Seek sets the offset for the next Read or Write operation on the underlying bytes.Reader.
// The offset is interpreted according to the whence parameter:
//   - io.SeekStart means relative to the start of the file
//   - io.SeekCurrent means relative to the current offset
//   - io.SeekEnd means relative to the end of the file
//
// Seek returns the new offset and any error encountered.
// It implements the io.Seeker interface.
func (b *bufferedFileReader) Seek(offset int64, whence int) (int64, error) {
	return b.reader.Seek(offset, whence)
}

// ReadAt reads len(p) bytes from the underlying io.ReadSeekCloser starting at byte offset off.
// It returns the number of bytes read and any error encountered.
// If the io.ReadSeekCloser reaches the end of the available data before len(p) bytes are read,
// ReadAt returns the number of bytes read and io.EOF.
// It implements the io.ReaderAt interface.
func (b *bufferedFileReader) ReadAt(p []byte, off int64) (n int, err error) {
	_, err = b.reader.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return b.reader.Read(p)
}
