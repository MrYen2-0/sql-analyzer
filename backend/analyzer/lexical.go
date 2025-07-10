package analyzer

import (
	"fmt"
	"regexp"
	"strings"
)

type Token struct {
	Type  string `json:"tipo"`
	Value string `json:"token"`
}

var keywords = []string{
	"SELECT", "FROM", "WHERE", "INSERT", "INTO", "VALUES",
	"UPDATE", "SET", "DELETE", "CREATE", "TABLE", "DROP",
	"ALTER", "ADD", "COLUMN", "PRIMARY", "KEY", "FOREIGN",
	"REFERENCES", "NOT", "NULL", "UNIQUE", "DEFAULT",
	"AND", "OR", "IN", "BETWEEN", "LIKE", "ORDER", "BY",
	"GROUP", "HAVING", "JOIN", "INNER", "LEFT", "RIGHT",
	"ON", "AS", "DISTINCT", "LIMIT", "OFFSET", "UNION",
	"ALL", "DATABASE", "USE", "IF", "EXISTS", "CASCADE",
	"CONSTRAINT", "INDEX", "VIEW", "PROCEDURE", "FUNCTION",
	"TRIGGER", "BEGIN", "END", "COMMIT", "ROLLBACK",
}

func LexicalAnalysis(query string) ([]Token, error) {
	var tokens []Token
	query = strings.TrimSpace(query)

	if query == "" {
		return nil, fmt.Errorf("query vacía")
	}

	// Patrones de expresiones regulares
	patterns := map[string]*regexp.Regexp{
		"NUMBER":     regexp.MustCompile(`^\d+(\.\d+)?`),
		"STRING":     regexp.MustCompile(`^'[^']*'`),
		"IDENTIFIER": regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*`),
		"OPERATOR":   regexp.MustCompile(`^(>=|<=|<>|!=|[><=+\-*/])`),
		"DELIMITER":  regexp.MustCompile(`^[(),;.]`),
	}

	i := 0
	for i < len(query) {
		// Saltar espacios en blanco
		if query[i] == ' ' || query[i] == '\t' || query[i] == '\n' {
			i++
			continue
		}

		matched := false
		remaining := query[i:]

		// Verificar operadores primero
		if match := patterns["OPERATOR"].FindString(remaining); match != "" {
			tokens = append(tokens, Token{Type: "OPERADOR", Value: match})
			i += len(match)
			matched = true
		} else if match := patterns["NUMBER"].FindString(remaining); match != "" {
			tokens = append(tokens, Token{Type: "NUMERO", Value: match})
			i += len(match)
			matched = true
		} else if match := patterns["STRING"].FindString(remaining); match != "" {
			tokens = append(tokens, Token{Type: "CADENA", Value: match})
			i += len(match)
			matched = true
		} else if match := patterns["IDENTIFIER"].FindString(remaining); match != "" {
			upperMatch := strings.ToUpper(match)
			tokenType := "IDENTIFICADOR"

			// Verificar si es una palabra clave
			for _, kw := range keywords {
				if upperMatch == kw {
					tokenType = "PALABRA_CLAVE"
					break
				}
			}

			tokens = append(tokens, Token{Type: tokenType, Value: match})
			i += len(match)
			matched = true
		} else if match := patterns["DELIMITER"].FindString(remaining); match != "" {
			tokens = append(tokens, Token{Type: "DELIMITADOR", Value: match})
			i += len(match)
			matched = true
		}

		if !matched {
			return nil, fmt.Errorf("carácter no reconocido: '%c' en posición %d", query[i], i)
		}
	}

	return tokens, nil
}
