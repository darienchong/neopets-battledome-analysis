package helpers

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Table struct {
	Name              string
	headers           []any
	rows              [][]any
	IsLastRowDistinct bool
}

func (table *Table) getColMaxSize(colIndex int) int {
	currMaxSize := len(table.headers[colIndex].(string))
	for _, row := range table.rows {
		currMaxSize = int(math.Max(float64(currMaxSize), float64(len(row[colIndex].(string)))))
	}
	return currMaxSize
}

func (table *Table) getLoggingTemplate() string {
	template := ""
	colMaxSizes := make([]int, len(table.headers))
	for i, _ := range table.headers {
		colMaxSizes[i] = table.getColMaxSize(i)
	}

	for i := 0; i < len(table.headers); i++ {
		template += "| %-" + strconv.Itoa(colMaxSizes[i]) + "s "
		if i == len(table.headers)-1 {
			template += "|"
		}
	}

	return template
}

func NewNamedTable(name string, headers []string) *Table {
	table := NewTable(headers)
	table.Name = name
	return table
}

func NewTable(headers []string) *Table {
	anyHeaders := make([]any, len(headers))
	for i, header := range headers {
		anyHeaders[i] = header
	}
	table := &Table{
		headers: anyHeaders,
		rows:    make([][]any, 0),
	}
	return table
}

func (table *Table) AddRow(row []string) error {
	if len(row) != len(table.headers) {
		return fmt.Errorf("row length did not match header length! row length was %d, but header length was %d", len(row), len(table.headers))
	}

	anyRow := make([]any, len(row))
	for i, val := range row {
		anyRow[i] = val
	}
	table.rows = append(table.rows, anyRow)
	return nil
}

func (table *Table) generateTableLineWithoutColumnSeparators() string {
	separator := ""
	for i := range table.headers {
		if i == 0 {
			separator += "|"
		} else {
			separator += "="
		}
		separator += strings.Repeat("=", table.getColMaxSize(i)+2)
		if i == len(table.headers)-1 {
			separator += "|"
		}
	}
	return separator
}

func (table *Table) generateTableLine() string {
	separator := ""
	for i := range table.headers {
		separator += "|"
		separator += strings.Repeat("=", table.getColMaxSize(i)+2)
		if i == len(table.headers)-1 {
			separator += "|"
		}
	}
	return separator
}

func (table *Table) GetLines() []string {
	lines := []string{}

	template := table.getLoggingTemplate()

	if table.Name != "" {
		rowSeparator := table.generateTableLineWithoutColumnSeparators()
		numSpaces := len(" ") + len(rowSeparator) - len(table.Name) - 3
		lines = append(lines, rowSeparator)
		leftPadding := int(math.Floor(float64(numSpaces) / 2.0))
		rightPadding := int(math.Ceil(float64(numSpaces) / 2.0))
		lines = append(lines, "|"+strings.Repeat(" ", leftPadding)+table.Name+strings.Repeat(" ", rightPadding)+"|")
	}

	lines = append(lines, table.generateTableLine())
	lines = append(lines, fmt.Sprintf(template, table.headers...))
	lines = append(lines, table.generateTableLine())
	for i, row := range table.rows {
		if table.IsLastRowDistinct && i == len(table.rows)-1 {
			lines = append(lines, table.generateTableLine())
		}
		lines = append(lines, fmt.Sprintf(template, row...))
	}
	lines = append(lines, table.generateTableLine())
	return lines
}

func getEmptyLine(length int) string {
	return strings.Repeat(" ", length)
}

func (table *Table) GetLinesWith(tableSeparator string, tables ...*Table) []string {
	tables = append([]*Table{table}, tables...)
	lines := []string{}

	tableLines := Map(tables, func(table *Table) []string {
		return table.GetLines()
	})

	isNamedTableExists := false
	for _, table := range tables {
		if table.Name != "" {
			isNamedTableExists = true
			break
		}
	}

	maxTableLength := Max(
		Map(
			tableLines,
			func(tableLines []string) int {
				return len(tableLines)
			},
		),
		func(a int, b int) bool {
			return a < b
		},
	)

	for i := range maxTableLength {
		lineParts := []string{}
		for j := range tables {
			currTable := tables[j]
			currTableLines := tableLines[j]
			if i < 2 && isNamedTableExists && currTable.Name == "" {
				lineParts = append(lineParts, getEmptyLine(len(currTableLines[0])))
				continue
			}

			if currTable.Name == "" {
				if 0 <= i-2 && i-2 < len(currTableLines) {
					lineParts = append(lineParts, currTableLines[i-2])
				} else {
					lineParts = append(lineParts, getEmptyLine(len(currTableLines[0])))
				}
			} else {
				if i < len(currTableLines) {
					lineParts = append(lineParts, currTableLines[i])
				} else {
					lineParts = append(lineParts, getEmptyLine(len(currTableLines[0])))
				}
			}
		}

		lines = append(lines, strings.Join(lineParts, tableSeparator))
	}

	return lines
}
