package backend

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/luismiguelgilolivert/gcs-emulator/internal/bigquery/model"
)

var (
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrTableNotFound   = errors.New("table not found")
)

type storedTable struct {
	meta *model.Table
	rows [][]interface{}
}

type MemoryBigQueryBackend struct {
	mu       sync.RWMutex
	datasets map[string]*model.Dataset
	tables   map[string]*storedTable
	jobID    int64
}

func NewMemoryBigQueryBackend() *MemoryBigQueryBackend {
	return &MemoryBigQueryBackend{
		datasets: make(map[string]*model.Dataset),
		tables:   make(map[string]*storedTable),
	}
}

func dsKey(project, name string) string { return project + ":" + name }
func tblKey(project, ds, name string) string { return project + ":" + ds + "." + name }

func (m *MemoryBigQueryBackend) CreateDataset(ctx context.Context, project string, ds *model.Dataset) (*model.Dataset, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	key := dsKey(project, ds.DatasetReference.DatasetID)
	if _, exists := m.datasets[key]; exists {
		return nil, errors.New("dataset already exists")
	}
	ds.Kind = "bigquery#dataset"
	m.datasets[key] = ds
	return ds, nil
}

func (m *MemoryBigQueryBackend) GetDataset(ctx context.Context, project, dataset string) (*model.Dataset, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	ds, exists := m.datasets[dsKey(project, dataset)]
	if !exists { return nil, ErrDatasetNotFound }
	return ds, nil
}

func (m *MemoryBigQueryBackend) ListDatasets(ctx context.Context, project string) ([]*model.Dataset, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Dataset
	for _, ds := range m.datasets {
		if ds.DatasetReference.ProjectID == project {
			result = append(result, ds)
		}
	}
	return result, nil
}

func (m *MemoryBigQueryBackend) DeleteDataset(ctx context.Context, project, dataset string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	key := dsKey(project, dataset)
	if _, exists := m.datasets[key]; !exists { return ErrDatasetNotFound }
	delete(m.datasets, key)
	for k := range m.tables {
		if strings.HasPrefix(k, project+":"+dataset+".") {
			delete(m.tables, k)
		}
	}
	return nil
}

func (m *MemoryBigQueryBackend) CreateTable(ctx context.Context, project, dataset string, table *model.Table) (*model.Table, error) {
	m.mu.Lock(); defer m.mu.Unlock()
	if _, exists := m.datasets[dsKey(project, dataset)]; !exists {
		return nil, ErrDatasetNotFound
	}
	key := tblKey(project, dataset, table.TableReference.TableID)
	if _, exists := m.tables[key]; exists {
		return nil, errors.New("table already exists")
	}
	table.Kind = "bigquery#table"
	m.tables[key] = &storedTable{meta: table, rows: make([][]interface{}, 0)}
	return table, nil
}

func (m *MemoryBigQueryBackend) GetTable(ctx context.Context, project, dataset, table string) (*model.Table, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	t, exists := m.tables[tblKey(project, dataset, table)]
	if !exists { return nil, ErrTableNotFound }
	return t.meta, nil
}

func (m *MemoryBigQueryBackend) ListTables(ctx context.Context, project, dataset string) ([]*model.Table, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	var result []*model.Table
	prefix := project + ":" + dataset + "."
	for k, t := range m.tables {
		if strings.HasPrefix(k, prefix) {
			result = append(result, t.meta)
		}
	}
	return result, nil
}

func (m *MemoryBigQueryBackend) DeleteTable(ctx context.Context, project, dataset, table string) error {
	m.mu.Lock(); defer m.mu.Unlock()
	key := tblKey(project, dataset, table)
	if _, exists := m.tables[key]; !exists { return ErrTableNotFound }
	delete(m.tables, key)
	return nil
}

func (m *MemoryBigQueryBackend) InsertRows(ctx context.Context, project, dataset, table string, rows []*model.Row) error {
	m.mu.Lock(); defer m.mu.Unlock()
	t, exists := m.tables[tblKey(project, dataset, table)]
	if !exists { return ErrTableNotFound }
	for _, row := range rows {
		data := make([]interface{}, len(row.F))
		for i, cell := range row.F {
			data[i] = cell.V
		}
		t.rows = append(t.rows, data)
	}
	t.meta.NumRows = int64(len(t.rows))
	return nil
}

func (m *MemoryBigQueryBackend) Query(ctx context.Context, query string) (*model.QueryResponse, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	return executeQuery(query, m.tables)
}

func executeQuery(sql string, tables map[string]*storedTable) (*model.QueryResponse, error) {
	sql = strings.TrimSpace(sql)
	upperSQL := strings.ToUpper(sql)

	if !strings.HasPrefix(upperSQL, "SELECT") {
		return nil, errors.New("only SELECT queries are supported")
	}

	fromIdx := strings.Index(upperSQL, "FROM ")
	if fromIdx < 0 {
		return nil, errors.New("missing FROM clause")
	}

	selectPart := strings.TrimSpace(sql[len("SELECT"):fromIdx])
	rest := strings.TrimSpace(sql[fromIdx+5:])

	whereClause := ""
	whereIdx := strings.Index(strings.ToUpper(rest), " WHERE ")
	if whereIdx >= 0 {
		whereClause = strings.TrimSpace(rest[whereIdx+7:])
		limitIdx := strings.Index(strings.ToUpper(whereClause), " LIMIT ")
		if limitIdx >= 0 {
			whereClause = strings.TrimSpace(whereClause[:limitIdx])
		}
	}

	limitIdx := strings.Index(strings.ToUpper(rest), " LIMIT ")
	limit := 100
	if limitIdx >= 0 {
		l := strings.TrimSpace(rest[limitIdx+7:])
		l = strings.SplitN(l, " ", 2)[0]
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}

	tableName := rest
	if whereIdx >= 0 {
		tableName = strings.TrimSpace(rest[:whereIdx])
	} else if limitIdx >= 0 {
		tableName = strings.TrimSpace(rest[:limitIdx])
	}
	tableName = strings.TrimSuffix(tableName, ";")
	tableName = strings.TrimSpace(tableName)

	t, exists := tables[tableName]
	if !exists {
		for k, v := range tables {
			if strings.HasSuffix(k, "."+tableName) || k == tableName {
				t = v
				exists = true
				break
			}
		}
	}
	if !exists {
		return nil, fmt.Errorf("table not found: %s", tableName)
	}

	schema := t.meta.Schema
	var selectedCols []int
	var colNames []string
	if selectPart == "*" {
		for i, f := range schema.Fields {
			selectedCols = append(selectedCols, i)
			colNames = append(colNames, f.Name)
		}
	} else {
		for _, part := range strings.Split(selectPart, ",") {
			colName := strings.TrimSpace(part)
			found := false
			for i, f := range schema.Fields {
				if strings.EqualFold(f.Name, colName) {
					selectedCols = append(selectedCols, i)
					colNames = append(colNames, f.Name)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("column not found: %s", colName)
			}
		}
	}

	var filteredRows [][]interface{}
	for _, row := range t.rows {
		if matchWhere(row, schema.Fields, whereClause) {
			filteredRows = append(filteredRows, row)
		}
	}

	var fields []*model.Field
	var resultRows []*model.Row

	for _, colIdx := range selectedCols {
		fields = append(fields, &model.Field{
			Name: schema.Fields[colIdx].Name,
			Type: schema.Fields[colIdx].Type,
		})
	}

	count := 0
	for _, row := range filteredRows {
		if count >= limit { break }
		r := &model.Row{F: make([]*model.Cell, len(selectedCols))}
		for j, colIdx := range selectedCols {
			r.F[j] = &model.Cell{V: row[colIdx]}
		}
		resultRows = append(resultRows, r)
		count++
	}

	return &model.QueryResponse{
		Kind:         "bigquery#queryResponse",
		Schema:       &model.TableSchema{Fields: fields},
		Rows:         resultRows,
		TotalRows:    int64(len(filteredRows)),
		JobComplete:  true,
	}, nil
}

func matchWhere(row []interface{}, fields []*model.Field, where string) bool {
	if where == "" { return true }
	upper := strings.ToUpper(where)

	parts := strings.SplitN(upper, "=", 2)
	if len(parts) == 2 {
		colName := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
		colIdx := -1
		for i, f := range fields {
			if strings.EqualFold(f.Name, colName) {
				colIdx = i
				break
			}
		}
		if colIdx >= 0 {
			rowVal := fmt.Sprintf("%v", row[colIdx])
			return strings.EqualFold(rowVal, val)
		}
	}

	parts = strings.SplitN(upper, ">", 2)
	if len(parts) == 2 {
		colName := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		colIdx := -1
		for i, f := range fields {
			if strings.EqualFold(f.Name, colName) {
				colIdx = i
				break
			}
		}
		if colIdx >= 0 {
			rowNum := toFloat(row[colIdx])
			filterNum, _ := strconv.ParseFloat(val, 64)
			return rowNum > filterNum
		}
	}

	return true
}

func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64: return val
	case int: return float64(val)
	case int64: return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	}
	return 0
}

func (m *MemoryBigQueryBackend) Shutdown() error { return nil }

var _ = sort.Ints
