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

package convert

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/apache/arrow/go/v17/arrow"
	filesystem "github.com/arrowarc/arrowarc/internal/integrations/filesystem"
)

func ConvertParquetToCSV(ctx context.Context, parquetFilePath, csvFilePath string, memoryMap bool, chunkSize int64, columns []string, rowGroups []int, parallel bool, delimiter rune, includeHeader bool, nullValue string, stringsReplacer *strings.Replacer, boolFormatter func(bool) string) error {
	recordChan, errChan := filesystem.ReadParquetFileStream(ctx, parquetFilePath, memoryMap, chunkSize, columns, rowGroups, parallel)

	var schema *arrow.Schema
	for rec := range recordChan {
		schema = rec.Schema()
		break
	}

	if schema == nil {
		return fmt.Errorf("could not determine schema from Parquet file")
	}

	writeErrChan := filesystem.WriteCSVFileStream(ctx, csvFilePath, schema, recordChan, delimiter, includeHeader, nullValue, stringsReplacer, boolFormatter)

	var wg sync.WaitGroup
	wg.Add(2)

	var readErr, writeErr error

	go func() {
		defer wg.Done()
		for err := range errChan {
			if err != nil {
				readErr = fmt.Errorf("error while reading Parquet file: %w", err)
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		for err := range writeErrChan {
			if err != nil {
				writeErr = fmt.Errorf("error while writing CSV file: %w", err)
				return
			}
		}
	}()

	wg.Wait()

	if readErr != nil {
		return readErr
	}
	if writeErr != nil {
		return writeErr
	}
	return nil
}