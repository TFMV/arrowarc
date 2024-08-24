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

package integrations

import (
	"context"
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	duckdb "github.com/arrowarc/arrowarc/internal/integrations/duckdb"
)

func ReadIcebergFileStream(ctx context.Context, filePath string) (<-chan arrow.Record, <-chan error) {
	extensions := []duckdb.DuckDBExtension{
		{Name: "httpfs", LoadByDefault: true},
		{Name: "iceberg", LoadByDefault: true},
	}

	conn, err := duckdb.OpenDuckDBConnection(ctx, "", extensions)
	if err != nil {
		errChan := make(chan error, 1)
		close(errChan)
		return nil, errChan
	}

	go func() {
		<-ctx.Done()
		duckdb.CloseDuckDBConnection(conn)
	}()

	query := fmt.Sprintf("SELECT * FROM iceberg_scan('%s')", filePath)

	return duckdb.ReadDuckDBStream(ctx, conn, query)
}