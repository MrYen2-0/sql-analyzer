package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	var err error
	connStr := "host=localhost port=5432 user=postgres password=yuen dbname=testdb sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error conectando a PostgreSQL:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Error verificando conexión a PostgreSQL:", err)
	}

	log.Println("Conectado a PostgreSQL")
}

func GetDB() *sql.DB {
	return db
}

type QueryResult struct {
	Type         string                   `json:"type"`
	RowsAffected int64                    `json:"rowsAffected,omitempty"`
	Data         []map[string]interface{} `json:"data,omitempty"`
	Columns      []string                 `json:"columns,omitempty"`
	Message      string                   `json:"message"`
	TableName    string                   `json:"tableName,omitempty"`
}

func ExecuteQuery(query string) (*QueryResult, error) {
	queryUpper := strings.ToUpper(strings.TrimSpace(query))

	switch {
	case strings.HasPrefix(queryUpper, "SELECT"):
		return executeSelect(query)
	case strings.HasPrefix(queryUpper, "INSERT"):
		return executeInsert(query)
	case strings.HasPrefix(queryUpper, "UPDATE"):
		return executeUpdate(query)
	case strings.HasPrefix(queryUpper, "DELETE"):
		return executeDelete(query)
	case strings.HasPrefix(queryUpper, "CREATE"):
		return executeCreate(query)
	case strings.HasPrefix(queryUpper, "DROP"):
		return executeDrop(query)
	default:
		return executeGeneric(query)
	}
}

func executeSelect(query string) (*QueryResult, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	return &QueryResult{
		Type:    "SELECT",
		Data:    results,
		Columns: columns,
		Message: fmt.Sprintf("Consulta ejecutada. %d filas encontradas.", len(results)),
	}, nil
}

func executeInsert(query string) (*QueryResult, error) {
	// Extraer nombre de tabla
	tableName := extractTableName(query, "INTO")

	result, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()

	// Obtener los datos insertados (últimas N filas)
	var selectQuery string
	if tableName != "" {
		selectQuery = fmt.Sprintf("SELECT * FROM %s ORDER BY id DESC LIMIT %d", tableName, rowsAffected)
		rows, err := db.Query(selectQuery)
		if err == nil {
			defer rows.Close()
			data, columns := rowsToData(rows)

			return &QueryResult{
				Type:         "INSERT",
				RowsAffected: rowsAffected,
				Data:         data,
				Columns:      columns,
				Message:      fmt.Sprintf("INSERT exitoso. %d fila(s) insertada(s) en %s.", rowsAffected, tableName),
				TableName:    tableName,
			}, nil
		}
	}

	return &QueryResult{
		Type:         "INSERT",
		RowsAffected: rowsAffected,
		Message:      fmt.Sprintf("INSERT exitoso. %d fila(s) insertada(s).", rowsAffected),
		TableName:    tableName,
	}, nil
}

func executeUpdate(query string) (*QueryResult, error) {
	tableName := extractTableName(query, "UPDATE")

	// Ejecutar el UPDATE directamente
	result, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()

	// Obtener los datos después del update
	if tableName != "" && rowsAffected > 0 {
		whereClause := extractWhereClause(query)
		selectQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		rows, err := db.Query(selectQuery)
		if err == nil {
			defer rows.Close()
			data, columns := rowsToData(rows)

			return &QueryResult{
				Type:         "UPDATE",
				RowsAffected: rowsAffected,
				Data:         data,
				Columns:      columns,
				Message:      fmt.Sprintf("UPDATE exitoso. %d fila(s) actualizada(s) en %s.", rowsAffected, tableName),
				TableName:    tableName,
			}, nil
		}
	}

	return &QueryResult{
		Type:         "UPDATE",
		RowsAffected: rowsAffected,
		Message:      fmt.Sprintf("UPDATE exitoso. %d fila(s) actualizada(s).", rowsAffected),
		TableName:    tableName,
	}, nil
}

func executeDelete(query string) (*QueryResult, error) {
	tableName := extractTableName(query, "FROM")

	// Obtener datos antes de eliminar
	var deletedData []map[string]interface{}
	var columns []string
	if tableName != "" {
		whereClause := extractWhereClause(query)
		selectQuery := fmt.Sprintf("SELECT * FROM %s %s", tableName, whereClause)
		rows, err := db.Query(selectQuery)
		if err == nil {
			deletedData, columns = rowsToData(rows)
			rows.Close()
		}
	}

	result, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()

	return &QueryResult{
		Type:         "DELETE",
		RowsAffected: rowsAffected,
		Data:         deletedData,
		Columns:      columns,
		Message:      fmt.Sprintf("DELETE exitoso. %d fila(s) eliminada(s) de %s.", rowsAffected, tableName),
		TableName:    tableName,
	}, nil
}

func executeCreate(query string) (*QueryResult, error) {
	_, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	queryUpper := strings.ToUpper(query)
	var objectType, objectName string

	if strings.Contains(queryUpper, "TABLE") {
		objectType = "TABLA"
		objectName = extractTableName(query, "TABLE")
	} else if strings.Contains(queryUpper, "DATABASE") {
		objectType = "BASE DE DATOS"
		objectName = extractTableName(query, "DATABASE")
	} else if strings.Contains(queryUpper, "INDEX") {
		objectType = "ÍNDICE"
		objectName = extractTableName(query, "INDEX")
	}

	// Si es una tabla, obtener su estructura
	if objectType == "TABLA" && objectName != "" {
		structQuery := `
            SELECT column_name, data_type, is_nullable, column_default
            FROM information_schema.columns
            WHERE table_name = $1
            ORDER BY ordinal_position;
        `
		rows, err := db.Query(structQuery, strings.ToLower(objectName))
		if err == nil {
			defer rows.Close()
			data, columns := rowsToData(rows)

			return &QueryResult{
				Type:      "CREATE",
				Data:      data,
				Columns:   columns,
				Message:   fmt.Sprintf("%s '%s' creada exitosamente.", objectType, objectName),
				TableName: objectName,
			}, nil
		}
	}

	return &QueryResult{
		Type:    "CREATE",
		Message: fmt.Sprintf("%s creado exitosamente.", objectType),
	}, nil
}

func executeDrop(query string) (*QueryResult, error) {
	queryUpper := strings.ToUpper(query)
	var objectType, objectName string

	if strings.Contains(queryUpper, "TABLE") {
		objectType = "TABLA"
		objectName = extractTableName(query, "TABLE")
	} else if strings.Contains(queryUpper, "DATABASE") {
		objectType = "BASE DE DATOS"
		objectName = extractTableName(query, "DATABASE")
	}

	_, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	return &QueryResult{
		Type:      "DROP",
		Message:   fmt.Sprintf("%s '%s' eliminada exitosamente.", objectType, objectName),
		TableName: objectName,
	}, nil
}

func executeGeneric(query string) (*QueryResult, error) {
	result, err := db.Exec(query)
	if err != nil {
		return nil, err
	}

	rowsAffected, _ := result.RowsAffected()

	return &QueryResult{
		Type:         "QUERY",
		RowsAffected: rowsAffected,
		Message:      "Query ejecutada exitosamente.",
	}, nil
}

// Funciones auxiliares
func extractTableName(query string, afterKeyword string) string {
	parts := strings.Fields(strings.ToUpper(query))
	for i, part := range parts {
		if part == afterKeyword && i+1 < len(parts) {
			tableName := parts[i+1]
			// Limpiar el nombre de la tabla
			tableName = strings.Trim(tableName, "(")
			tableName = strings.ToLower(tableName)
			return tableName
		}
	}
	return ""
}

func extractWhereClause(query string) string {
	upperQuery := strings.ToUpper(query)
	whereIndex := strings.Index(upperQuery, "WHERE")
	if whereIndex != -1 {
		return query[whereIndex:]
	}
	return ""
}

func rowsToData(rows *sql.Rows) ([]map[string]interface{}, []string) {
	columns, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			continue
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	return results, columns
}

// Función para obtener el estado actual de todas las tablas
func GetDatabaseState() (map[string]interface{}, error) {
	tablesQuery := `
        SELECT table_name 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        ORDER BY table_name;
    `

	rows, err := db.Query(tablesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	state := make(map[string]interface{})
	var tables []map[string]interface{}

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}

		// Contar filas en cada tabla
		var count int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
		db.QueryRow(countQuery).Scan(&count)

		tableInfo := map[string]interface{}{
			"name":     tableName,
			"rowCount": count,
		}
		tables = append(tables, tableInfo)
	}

	state["tables"] = tables
	state["totalTables"] = len(tables)

	return state, nil
}
