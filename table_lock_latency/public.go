// Package table_lock_latency represents the performance_schema table of the same name
package table_lock_latency

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql" // keep golint happy
	"time"

	"github.com/sjmudd/ps-top/logger"
	"github.com/sjmudd/ps-top/p_s"
)

const (
	description = "Locks by Table Name (table_lock_waits_summary_by_table)"
)

// Object represents a table of rows
type Object struct {
	p_s.RelativeStats
	p_s.CollectionTime
	initial Rows // initial data for relative values
	current Rows // last loaded values
	results Rows // results (maybe with subtraction)
	totals  Row  // totals of results
}

// Collect data from the db, then merge it in.
func (t *Object) Collect(dbh *sql.DB) {
	start := time.Now()
	t.current = selectRows(dbh)

	if len(t.initial) == 0 && len(t.current) > 0 {
		t.initial = make(Rows, len(t.current))
		copy(t.initial, t.current)
	}

	// check for reload initial characteristics
	if t.initial.needsRefresh(t.current) {
		t.initial = make(Rows, len(t.current))
		copy(t.initial, t.current)
	}

	t.makeResults()
	logger.Println("Object.Collect() took:", time.Duration(time.Since(start)).String())
}

func (t *Object) makeResults() {
	// logger.Println( "- t.results set from t.current" )
	t.results = make(Rows, len(t.current))
	copy(t.results, t.current)
	if t.WantRelativeStats() {
		// logger.Println( "- subtracting t.initial from t.results as WantRelativeStats()" )
		t.results.subtract(t.initial)
	}

	// logger.Println( "- sorting t.results" )
	t.results.sort()
	// logger.Println( "- collecting t.totals from t.results" )
	t.totals = t.results.totals()
}

// SetInitialFromCurrent resets the statistics to current values
func (t *Object) SetInitialFromCurrent() {
	t.SetCollected()
	t.initial = make(Rows, len(t.current))
	copy(t.initial, t.current)

	t.makeResults()
}

// Headings returns the headings for a table
func (t Object) Headings() string {
	var r Row

	return r.headings()
}

// RowContent returns the rows we need for displaying
func (t Object) RowContent(maxRows int) []string {
	rows := make([]string, 0, maxRows)

	for i := range t.results {
		if i < maxRows {
			rows = append(rows, t.results[i].rowContent(t.totals))
		}
	}

	return rows
}

// TotalRowContent returns all the totals
func (t Object) TotalRowContent() string {
	return t.totals.rowContent(t.totals)
}

// EmptyRowContent returns an empty string of data (for filling in)
func (t Object) EmptyRowContent() string {
	var emtpy Row
	return emtpy.rowContent(emtpy)
}

// Description provides a description of the table
func (t Object) Description() string {
	return description
}

// Len returns the length of the result set
func (t Object) Len() int {
	return len(t.results)
}
