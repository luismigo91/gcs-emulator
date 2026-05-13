package model

type Dataset struct {
	Kind        string            `json:"kind"`
	ID          string            `json:"id"`
	DatasetReference *DatasetReference `json:"datasetReference"`
	FriendlyName string           `json:"friendlyName,omitempty"`
	Location    string            `json:"location,omitempty"`
}

type DatasetReference struct {
	DatasetID string `json:"datasetId"`
	ProjectID string `json:"projectId"`
}

type Table struct {
	Kind              string             `json:"kind"`
	ID                string             `json:"id"`
	TableReference    *TableReference    `json:"tableReference"`
	Schema            *TableSchema       `json:"schema,omitempty"`
	NumRows           int64              `json:"numRows,string,omitempty"`
	NumBytes          int64              `json:"numBytes,string,omitempty"`
}

type TableReference struct {
	DatasetID string `json:"datasetId"`
	ProjectID string `json:"projectId"`
	TableID   string `json:"tableId"`
}

type TableSchema struct {
	Fields []*Field `json:"fields"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Mode string `json:"mode,omitempty"`
}

type QueryRequest struct {
	Query          string `json:"query"`
	DefaultDataset *DatasetReference `json:"defaultDataset,omitempty"`
	MaxResults     int    `json:"maxResults,omitempty"`
}

type QueryResponse struct {
	Kind           string           `json:"kind"`
	JobReference   *JobReference    `json:"jobReference"`
	Schema         *TableSchema     `json:"schema,omitempty"`
	Rows           []*Row           `json:"rows,omitempty"`
	TotalRows      int64            `json:"totalRows,string,omitempty"`
	JobComplete    bool             `json:"jobComplete"`
}

type JobReference struct {
	ProjectID string `json:"projectId"`
	JobID     string `json:"jobId"`
	Location  string `json:"location,omitempty"`
}

type Row struct {
	F []*Cell `json:"f"`
}

type Cell struct {
	V interface{} `json:"v"`
}

type InsertRequest struct {
	Rows []*Row `json:"rows"`
}
