package helpers

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/palantir/stacktrace"
)

type Table struct {
	Name              string
	headers           []any
	rows              [][]any
	IsLastRowDistinct bool
}

func (t *Table) colMaxSize(colIndex int) int {
	currMaxSize := len(t.headers[colIndex].(string))
	for _, row := range t.rows {
		currMaxSize = int(math.Max(float64(currMaxSize), float64(len(row[colIndex].(string)))))
	}
	return currMaxSize
}

func (t *Table) loggingTemplate() string {
	template := ""
	colMaxSizes := make([]int, len(t.headers))
	for i, _ := range t.headers {
		colMaxSizes[i] = t.colMaxSize(i)
	}

	for i := 0; i < len(t.headers); i++ {
		template += "| %-" + strconv.Itoa(colMaxSizes[i]) + "s "
		if i == len(t.headers)-1 {
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

func (t *Table) AddRow(row []string) error {
	if len(row) != len(t.headers) {
		return fmt.Errorf("row length did not match header length! row length was %d, but header length was %d", len(row), len(t.headers))
	}

	anyRow := make([]any, len(row))
	for i, val := range row {
		anyRow[i] = val
	}
	t.rows = append(t.rows, anyRow)
	return nil
}

func (t *Table) generateTableLineWithoutColumnSeparators() string {
	separator := ""
	for i := range t.headers {
		if i == 0 {
			separator += "|"
		} else {
			separator += "="
		}
		separator += strings.Repeat("=", t.colMaxSize(i)+2)
		if i == len(t.headers)-1 {
			separator += "|"
		}
	}
	return separator
}

func (t *Table) generateTableLine() string {
	separator := ""
	for i := range t.headers {
		separator += "|"
		separator += strings.Repeat("=", t.colMaxSize(i)+2)
		if i == len(t.headers)-1 {
			separator += "|"
		}
	}
	return separator
}

func (t *Table) Lines() []string {
	lines := []string{}

	template := t.loggingTemplate()

	if t.Name != "" {
		rowSeparator := t.generateTableLineWithoutColumnSeparators()
		if len(strings.TrimSpace(t.Name)) >= len(rowSeparator)-2 {
			t.Name = strings.TrimSpace(t.Name[:len(" ")+len(rowSeparator)-8]) + "..."
		}
		numSpaces := len(" ") + len(rowSeparator) - len(t.Name) - 3

		lines = append(lines, rowSeparator)
		leftPadding := int(math.Ceil(float64(numSpaces) / 2.0))
		rightPadding := int(math.Floor(float64(numSpaces) / 2.0))
		if leftPadding < 0 || rightPadding < 0 {
			errorMsg := fmt.Sprintf("generated a negative left/right padding value (left: %d, right: %d) for the table name! the table name was \"%s\" (%d chars), and the table width was %d chars", leftPadding, rightPadding, t.Name, len(t.Name), len(rowSeparator))
			panic(stacktrace.NewError(errorMsg))
		}
		lines = append(lines, "|"+strings.Repeat(" ", leftPadding)+t.Name+strings.Repeat(" ", rightPadding)+"|")
	}

	lines = append(lines, t.generateTableLine())
	lines = append(lines, fmt.Sprintf(template, t.headers...))
	lines = append(lines, t.generateTableLine())
	for i, row := range t.rows {
		if t.IsLastRowDistinct && i == len(t.rows)-1 {
			lines = append(lines, t.generateTableLine())
		}
		lines = append(lines, fmt.Sprintf(template, row...))
	}
	lines = append(lines, t.generateTableLine())
	return lines
}

func emptyLine(length int) string {
	return strings.Repeat(" ", length)
}

func (t *Table) LinesWith(tableSeparator string, tables ...*Table) []string {
	tables = append([]*Table{t}, tables...)
	lines := []string{}

	tableLines := Map(tables, func(table *Table) []string {
		return table.Lines()
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
				lineParts = append(lineParts, emptyLine(len(currTableLines[0])))
				continue
			}

			if currTable.Name == "" {
				if 0 <= i-2 && i-2 < len(currTableLines) {
					lineParts = append(lineParts, currTableLines[i-2])
				} else {
					lineParts = append(lineParts, emptyLine(len(currTableLines[0])))
				}
			} else {
				if i < len(currTableLines) {
					lineParts = append(lineParts, currTableLines[i])
				} else {
					lineParts = append(lineParts, emptyLine(len(currTableLines[0])))
				}
			}
		}

		lines = append(lines, strings.Join(lineParts, tableSeparator))
	}

	return lines
}
