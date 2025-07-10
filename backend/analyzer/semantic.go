package analyzer

import (
	"database/sql"
	"fmt"
	"sql-analyzer/database"
	"strings"
)

type SemanticInfo struct {
	Tables   []TableInfo  `json:"tables"`
	Columns  []ColumnInfo `json:"columns"`
	Warnings []string     `json:"warnings"`
	Valid    bool         `json:"valid"`
}

type TableInfo struct {
	Name   string `json:"name"`
	Exists bool   `json:"exists"`
}

type ColumnInfo struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Type   string `json:"type"`
	Exists bool   `json:"exists"`
}

func SemanticAnalysis(query string) (*SemanticInfo, error) {
	tokens, err := LexicalAnalysis(query)
	if err != nil {
		return nil, err
	}

	info := &SemanticInfo{
		Valid:    true,
		Warnings: []string{},
	}

	// Extraer tablas y columnas de la consulta
	tables := extractTables(tokens)
	columns := extractColumns(tokens)

	// Verificar existencia de tablas
	db := database.GetDB()
	for _, table := range tables {
		exists := checkTableExists(db, table)
		info.Tables = append(info.Tables, TableInfo{
			Name:   table,
			Exists: exists,
		})
		if !exists {
			info.Valid = false
			info.Warnings = append(info.Warnings,
				fmt.Sprintf("La tabla '%s' no existe", table))
		}
	}

	// Verificar columnas
	for _, col := range columns {
		if col != "*" {
			// Verificación simplificada
			info.Columns = append(info.Columns, ColumnInfo{
				Column: col,
				Exists: true, // Simplificado
			})
		}
	}

	return info, nil
}

func extractTables(tokens []Token) []string {
	tables := []string{}
	fromFound := false

	for i, token := range tokens {
		if strings.ToUpper(token.Value) == "FROM" ||
			strings.ToUpper(token.Value) == "INTO" ||
			strings.ToUpper(token.Value) == "UPDATE" {
			fromFound = true
			continue
		}

		if fromFound && token.Type == "IDENTIFICADOR" {
			tables = append(tables, token.Value)
			fromFound = false
		}

		if strings.ToUpper(token.Value) == "TABLE" && i+1 < len(tokens) {
			if tokens[i+1].Type == "IDENTIFICADOR" {
				tables = append(tables, tokens[i+1].Value)
			}
		}
	}

	return tables
}

func extractColumns(tokens []Token) []string {
	columns := []string{}
	selectFound := false

	for _, token := range tokens {
		if strings.ToUpper(token.Value) == "SELECT" {
			selectFound = true
			continue
		}

		if selectFound && strings.ToUpper(token.Value) == "FROM" {
			break
		}

		if selectFound && (token.Type == "IDENTIFICADOR" || token.Value == "*") {
			columns = append(columns, token.Value)
		}
	}

	return columns
}

func checkTableExists(db *sql.DB, tableName string) bool {
	var exists bool
	query := `
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = $1
        );`

	err := db.QueryRow(query, strings.ToLower(tableName)).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}
