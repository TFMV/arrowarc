// --------------------------------------------------------------------------------
// Author: Thomas F McGeehan V
//
// This file is part of a software project developed by Thomas F McGeehan V.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// For more information about the MIT License, please visit:
// https://opensource.org/licenses/MIT
//
// Acknowledgment appreciated but not required.
// --------------------------------------------------------------------------------

package converter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	integrations "github.com/arrowarc/arrowarc/integrations/filesystem"
	"github.com/arrowarc/arrowarc/pipeline"
)

func ConvertParquetToCSV(
	ctx context.Context,
	parquetFilePath, csvFilePath string,
	memoryMap bool, chunkSize int64,
	columns []string, rowGroups []int, parallel bool,
	delimiter rune, includeHeader bool,
	nullValue string, stringsReplacer *strings.Replacer,
	boolFormatter func(bool) string,
) (string, error) {
	// Validate input parameters
	if parquetFilePath == "" {
		return "", errors.New("parquet file path cannot be empty")
	}
	if csvFilePath == "" {
		return "", errors.New("CSV file path cannot be empty")
	}
	if chunkSize <= 0 {
		return "", errors.New("chunk size must be greater than zero")
	}
	if ctx == nil {
		return "", errors.New("context cannot be nil")
	}

	// Create Parquet reader
	reader, err := integrations.NewParquetReader(ctx, parquetFilePath, &integrations.ParquetReadOptions{
		MemoryMap: memoryMap,
		RowGroups: rowGroups,
		Parallel:  parallel,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Parquet reader for file '%s': %w", parquetFilePath, err)
	}
	defer func() {
		if cerr := reader.Close(); cerr != nil {
			err = fmt.Errorf("failed to close Parquet reader: %w", cerr)
		}
	}()

	// Create CSV writer
	writer, err := integrations.NewCSVWriter(ctx, csvFilePath, reader.Schema(), &integrations.CSVWriteOptions{
		Delimiter:       delimiter,
		IncludeHeader:   includeHeader,
		NullValue:       nullValue,
		StringsReplacer: stringsReplacer,
		BoolFormatter:   boolFormatter,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create CSV writer for file '%s': %w", csvFilePath, err)
	}
	defer func() {
		if cerr := writer.Close(); cerr != nil {
			err = fmt.Errorf("failed to close CSV writer: %w", cerr)
		}
	}()

	// Setup pipeline
	p := pipeline.NewDataPipeline(reader, writer)

	// Start the pipeline and wait for completion
	metrics, startErr := p.Start(ctx)
	if startErr != nil {
		return "", fmt.Errorf("failed to start conversion pipeline: %w", startErr)
	}

	// Wait for the pipeline to finish
	if pipelineErr := <-p.Done(); pipelineErr != nil {
		return "", fmt.Errorf("pipeline encountered an error: %w", pipelineErr)
	}

	return metrics, nil
}
