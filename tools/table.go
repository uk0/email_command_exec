package tools

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Table - Table structure.
type Table struct {
	Fields     []string
	Footer     map[string]string
	Rows       []map[string]string
	HideHead   bool // when true doesn't print header
	Markdown   bool
	fieldSizes map[string]int
}

// New - Creates a new table.
func New(fields []string) *Table {
	return &Table{
		Fields:     fields,
		Rows:       make([]map[string]string, 0),
		fieldSizes: make(map[string]int),
	}
}

// AddRow - Adds row to the table.
func (t *Table) AddRow(row map[string]interface{}) {
	newRow := make(map[string]string)
	for _, k := range t.Fields {
		v := row[k]
		// If is not nil format
		// else value is empty string
		var val string
		if v == nil {
			val = ""
		} else {
			val = fmt.Sprintf("%v", v)
		}

		newRow[k] = val
	}

	t.calculateSizes(newRow)

	if len(newRow) > 0 {
		t.Rows = append(t.Rows, newRow)
	}
}

// AddFooter - Adds footer to the table.
func (t *Table) AddFooter(footer map[string]string) {
	t.Footer = footer
}

// Print - Prints table.
func (t *Table) GetText() string {
	result := ""
	if len(t.Rows) == 0 && t.Footer == nil {
		return ""
	}

	t.calculateSizes(t.Footer)

	if !t.Markdown {
		result = result + t.printDash() + "\n"
	}

	if !t.HideHead {
		result = result + t.getHead() + "\n"
		result = result + t.printTableDash() + "\n"
	}

	for _, r := range t.Rows {
		result = result + t.rowString(r) + "\n"
		if !t.Markdown {
			result = result + t.printDash() + "\n"
		}
	}

	if t.Footer != nil {
		result = result + t.printTableDash() + "\n"
		result = result + t.rowString(t.Footer) + "\n"
		if !t.Markdown {
			result = result + t.printTableDash() + "\n"
		}
	}
	return result
}

// getHead - Returns table header containing fields names.
func (t *Table) getHead() string {
	s := "|"
	for _, name := range t.Fields {
		s += t.fieldString(name, strings.Title(name)) + "|"
	}
	return s
}

// rowString - Creates a string row.
func (t *Table) rowString(row map[string]string) string {
	s := "|"
	for _, name := range t.Fields {
		value := row[name]
		s += t.fieldString(name, value) + "|"
	}
	return s
}

// fieldString - Creates field value string.
func (t *Table) fieldString(name, value string) string {
	value = fmt.Sprintf(" %s ", value)
	spacesLeft := t.fieldSizes[name] - runewidth.StringWidth(value)
	if spacesLeft > 0 {
		for i := 0; i < spacesLeft; i++ {
			value += " "
		}
	}
	return value
}

// printTableDash - Prints table dash. Markdown or not depending on settings.
func (t *Table) printTableDash() string {
	if t.Markdown {
		return t.printMarkdownDash()
	} else {
		return t.printDash()
	}
}

// printDash - Prints dash (on top and header).
func (t *Table) printDash() string {
	s := "|"
	for i := 0; i < t.lineLength()-2; i++ {
		s += "-"
	}
	s += "|"
	return s
}

// printMarkdownDash - Prints dash in middle of table.
func (t *Table) printMarkdownDash() string {
	r := make(map[string]string)
	for _, name := range t.Fields {
		r[name] = strings.Repeat("-", t.fieldSizes[name]-2)
	}
	return t.rowString(r)
}

// lineLength - Counts size of table line length (with spaces etc.).
func (t *Table) lineLength() (sum int) {
	for _, l := range t.fieldSizes {
		sum += l + 1
	}
	return sum + 1
}

func (t *Table) calculateSizes(row map[string]string) {
	for _, k := range t.Fields {
		v, ok := row[k]
		if !ok {
			continue
		}

		vlen := runewidth.StringWidth(v)
		// align to field name length
		if klen := runewidth.StringWidth(k); vlen < klen {
			vlen = klen
		}
		vlen += 2 // + 2 spaces
		if t.fieldSizes[k] < vlen {
			t.fieldSizes[k] = vlen
		}
	}
}

func mapToRows(m map[string]interface{}) (rows []map[string]interface{}) {
	rows = []map[string]interface{}{}
	for key, value := range m {
		row := map[string]interface{}{}
		row["Key"] = strings.Title(key)
		row["Value"] = value
		rows = append(rows, row)
	}
	return
}
