package helpers

import (
	"fmt"
	"log/slog"
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

func (table *Table) Log() {
	table.LogWithPrefix("")
}

func (table *Table) LogWithPrefix(logPrefix string) {
	template := table.getLoggingTemplate()

	if table.Name != "" {
		rowSeparator := table.generateTableLineWithoutColumnSeparators()
		numSpaces := len(" ") + len(rowSeparator) - len(table.Name) - 3
		slog.Info(logPrefix + rowSeparator)
		leftPadding := int(math.Floor(float64(numSpaces) / 2.0))
		rightPadding := int(math.Ceil(float64(numSpaces) / 2.0))
		slog.Info(logPrefix + "|" + strings.Repeat(" ", leftPadding) + table.Name + strings.Repeat(" ", rightPadding) + "|")
	}

	slog.Info(logPrefix + table.generateTableLine())
	slog.Info(logPrefix + fmt.Sprintf(template, table.headers...))
	slog.Info(logPrefix + table.generateTableLine())
	for i, row := range table.rows {
		if table.IsLastRowDistinct && i == len(table.rows)-1 {
			slog.Info(logPrefix + table.generateTableLine())
		}
		slog.Info(logPrefix + fmt.Sprintf(template, row...))
	}
	slog.Info(logPrefix + table.generateTableLine())
}

func (table *Table) LogAlongside(otherTable *Table, tableSeparator string) {
	table.LogAlongsideWithPrefix(otherTable, tableSeparator, "")
}

func (table *Table) LogAlongsideWithPrefix(otherTable *Table, tableSeparator string, logPrefix string) {
	if len(table.rows) != len(otherTable.rows) {
		panic(fmt.Errorf("can't simultaneously log two tables of differing length"))
	}

	if table.Name == "" && otherTable.Name != "" || table.Name != "" && otherTable.Name == "" {
		panic(fmt.Errorf("can't simultaneously log a named table and an unnamed table"))
	}

	template := table.getLoggingTemplate()
	otherTemplate := otherTable.getLoggingTemplate()
	separator := table.generateTableLine() + tableSeparator + otherTable.generateTableLine()

	if table.Name != "" {
		rowSeparator := table.generateTableLineWithoutColumnSeparators()
		otherRowSeparator := otherTable.generateTableLineWithoutColumnSeparators()

		numSpaces := len(" ") + len(rowSeparator) - len(table.Name) - 3
		otherNumSpaces := len(" ") + len(otherRowSeparator) - len(otherTable.Name) - 3

		slog.Info(logPrefix + rowSeparator + tableSeparator + otherRowSeparator)
		leftPadding, rightPadding := int(math.Floor(float64(numSpaces)/2.0)), int(math.Ceil(float64(numSpaces)/2.0))
		otherLeftPadding, otherRightPadding := int(math.Floor(float64(otherNumSpaces)/2.0)), int(math.Ceil(float64(otherNumSpaces)/2.0))

		nameRow := "|" + strings.Repeat(" ", leftPadding) + table.Name + strings.Repeat(" ", rightPadding) + "|"
		otherNameRow := "|" + strings.Repeat(" ", otherLeftPadding) + otherTable.Name + strings.Repeat(" ", otherRightPadding) + "|"
		slog.Info(logPrefix + nameRow + tableSeparator + otherNameRow)
	}

	slog.Info(logPrefix + separator)
	slog.Info(logPrefix + fmt.Sprintf(template, table.headers...) + tableSeparator + fmt.Sprintf(otherTemplate, otherTable.headers...))
	slog.Info(logPrefix + separator)
	for i := range table.rows {
		row := table.rows[i]
		otherRow := otherTable.rows[i]

		if table.IsLastRowDistinct && i == len(table.rows)-1 {
			slog.Info(logPrefix + separator)
		}

		slog.Info(logPrefix + fmt.Sprintf(template, row...) + tableSeparator + fmt.Sprintf(otherTemplate, otherRow...))
	}
	slog.Info(logPrefix + separator)
}
