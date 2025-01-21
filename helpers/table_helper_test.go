package helpers

import (
	"fmt"
	"log/slog"
	"testing"
)

func generateTable(numRows int) *Table {
	table := NewTable([]string{
		"Col 1",
		"Col 2",
		"Col 3",
	})

	for i := range numRows {
		table.AddRow([]string{
			fmt.Sprintf("Row %d,1", i),
			fmt.Sprintf("Row %d,2", i),
			fmt.Sprintf("Row %d,3", i),
		})
	}
	return table
}

func generateNamedTable(name string, numRows int) *Table {
	table := NewNamedTable(name, []string{
		"Col 1",
		"Col 2",
		"Col 3",
	})
	for i := range numRows {
		table.AddRow([]string{
			fmt.Sprintf("Row %d,1", i),
			fmt.Sprintf("Row %d,2", i),
			fmt.Sprintf("Row %d,3", i),
		})
	}
	return table
}

func TestTable(t *testing.T) {
	table := generateTable(5)

	for _, line := range table.GetLines() {
		slog.Info(line)
	}
}

func TestTableWithDistinctLastRow(t *testing.T) {
	table := generateTable(6)
	table.IsLastRowDistinct = true

	for _, line := range table.GetLines() {
		slog.Info(line)
	}
}

func TestNamedTable(t *testing.T) {
	table := generateNamedTable("Name", 5)

	for _, line := range table.GetLines() {
		slog.Info(line)
	}
}

func TestMultipleTables(t *testing.T) {
	firstTable := generateNamedTable("First", 5)
	secondTable := generateNamedTable("Second", 5)
	thirdTable := generateNamedTable("Third", 5)

	for _, line := range firstTable.GetLinesWith(" ", secondTable, thirdTable) {
		slog.Info(line)
	}
}

func TestMultipleTablesWithDifferentLineCounts(t *testing.T) {
	firstTable := generateNamedTable("First", 5)
	secondTable := generateNamedTable("Second", 6)
	thirdTable := generateNamedTable("Third", 7)

	for _, line := range firstTable.GetLinesWith(" ", secondTable, thirdTable) {
		slog.Info(line)
	}
}

func TestMixedNamedAndUnnamedTablesWithDifferentLineCounts(t *testing.T) {
	firstTable := generateTable(5)
	secondTable := generateNamedTable("Second", 6)
	thirdTable := generateNamedTable("Third", 7)

	for _, line := range firstTable.GetLinesWith(" ", secondTable, thirdTable) {
		slog.Info(line)
	}
}
