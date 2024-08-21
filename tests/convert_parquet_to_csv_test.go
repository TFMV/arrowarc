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

package test

import (
	"context"
	"os"
	"testing"
	"time"

	converter "github.com/ArrowArc/ArrowArc/pkg/convert"
	generator "github.com/ArrowArc/ArrowArc/pkg/parquet"
	"github.com/stretchr/testify/assert"
)

func TestConvertParquetToCSV(t *testing.T) {
	t.Parallel() // Parallelize the top-level test

	// Generate a sample Parquet file for testing
	parquetFilePath := "sample_test.parquet"
	csvFilePathWithHeader := "output_test_with_header.csv"
	csvFilePathWithoutHeader := "output_test_without_header.csv"

	// Generate the Parquet file on the fly
	err := generator.GenerateParquetFile(parquetFilePath, 100*1024, false) // 100 KB, simple structure
	assert.NoError(t, err, "Error should be nil when generating Parquet file")

	// Use t.Cleanup to ensure the files are removed after all tests complete
	t.Cleanup(func() {
		os.Remove(parquetFilePath)
		os.Remove(csvFilePathWithHeader)
		os.Remove(csvFilePathWithoutHeader)
	})

	tests := []struct {
		parquetFilePath string
		csvFilePath     string
		memoryMap       bool
		chunkSize       int64
		columns         []string
		rowGroups       []int
		parallel        bool
		delimiter       rune
		includeHeader   bool
		nullValue       string
		description     string
	}{
		{
			parquetFilePath: parquetFilePath,
			csvFilePath:     csvFilePathWithHeader,
			memoryMap:       false,
			chunkSize:       1024,
			columns:         nil, // Read all columns
			rowGroups:       nil, // Read all row groups
			parallel:        true,
			delimiter:       ',',
			includeHeader:   true,
			nullValue:       "NULL",
			description:     "Convert Parquet to CSV with header",
		},
		{
			parquetFilePath: parquetFilePath,
			csvFilePath:     csvFilePathWithoutHeader,
			memoryMap:       true,
			chunkSize:       2048,
			columns:         []string{"id", "name"}, // Read specific columns
			rowGroups:       nil,                    // Read all row groups
			parallel:        false,
			delimiter:       ',',
			includeHeader:   false,
			nullValue:       "NULL",
			description:     "Convert Parquet to CSV without header",
		},
	}

	for _, test := range tests {
		test := test // capture range variable
		t.Run(test.description, func(t *testing.T) {
			t.Parallel() // Parallelize each subtest

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err := converter.ConvertParquetToCSV(ctx, test.parquetFilePath, test.csvFilePath, test.memoryMap, test.chunkSize, test.columns, test.rowGroups, test.parallel, test.delimiter, test.includeHeader, test.nullValue, nil, nil)
			assert.NoError(t, err, "Error should be nil when converting Parquet to CSV")

			// Ensure the CSV file was created
			_, err = os.Stat(test.csvFilePath)
			assert.NoError(t, err, "CSV file should be created")

			// Use t.Cleanup to ensure the CSV file is removed after each test completes
			t.Cleanup(func() {
				os.Remove(test.csvFilePath)
			})
		})
	}
}