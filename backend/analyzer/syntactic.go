package analyzer

import (
	"fmt"
	"strings"
)

type SyntaxNode struct {
	Type     string       `json:"type"`
	Value    string       `json:"value,omitempty"`
	Children []SyntaxNode `json:"children,omitempty"`
}

// Stack para verificar balance de paréntesis
type ParenthesisStack struct {
	items []string
}

func (s *ParenthesisStack) Push(item string) {
	s.items = append(s.items, item)
}

func (s *ParenthesisStack) Pop() (string, bool) {
	if len(s.items) == 0 {
		return "", false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *ParenthesisStack) IsEmpty() bool {
	return len(s.items) == 0
}

func SyntacticAnalysis(query string) (*SyntaxNode, error) {
	tokens, err := LexicalAnalysis(query)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("query vacía")
	}

	// Verificar balance de paréntesis en toda la query
	if err := checkParenthesisBalance(tokens); err != nil {
		return nil, err
	}

	// Verificar punto y coma al final
	if len(tokens) > 0 && tokens[len(tokens)-1].Value != ";" {
		// Nota: No es un error, pero es mejor práctica terminar con ;
	}

	upperQuery := strings.ToUpper(strings.TrimSpace(query))

	// Análisis por tipo de sentencia
	switch {
	case strings.HasPrefix(upperQuery, "SELECT"):
		return analyzeSelect(tokens)
	case strings.HasPrefix(upperQuery, "INSERT"):
		return analyzeInsert(tokens)
	case strings.HasPrefix(upperQuery, "UPDATE"):
		return analyzeUpdate(tokens)
	case strings.HasPrefix(upperQuery, "DELETE"):
		return analyzeDelete(tokens)
	case strings.HasPrefix(upperQuery, "CREATE"):
		return analyzeCreate(tokens)
	case strings.HasPrefix(upperQuery, "DROP"):
		return analyzeDrop(tokens)
	default:
		return nil, fmt.Errorf("tipo de sentencia no reconocida: %s", tokens[0].Value)
	}
}

func checkParenthesisBalance(tokens []Token) error {
	stack := &ParenthesisStack{}

	for i, token := range tokens {
		if token.Value == "(" {
			stack.Push("(")
		} else if token.Value == ")" {
			if _, ok := stack.Pop(); !ok {
				return fmt.Errorf("paréntesis de cierre ')' sin paréntesis de apertura correspondiente en posición %d", i)
			}
		}
	}

	if !stack.IsEmpty() {
		return fmt.Errorf("paréntesis de apertura '(' sin cerrar")
	}

	return nil
}

func analyzeCreateTable(tokens []Token, startIndex int, root *SyntaxNode) (*SyntaxNode, error) {
	i := startIndex + 1

	// Verificar IF NOT EXISTS (opcional)
	ifNotExists := false
	if i+2 < len(tokens) &&
		strings.ToUpper(tokens[i].Value) == "IF" &&
		strings.ToUpper(tokens[i+1].Value) == "NOT" &&
		strings.ToUpper(tokens[i+2].Value) == "EXISTS" {
		ifNotExists = true
		i += 3
	}

	// Nombre de tabla
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de CREATE TABLE")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de tabla inválido: '%s'", tokens[i].Value)
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	if ifNotExists {
		tableNode.Children = append(tableNode.Children,
			SyntaxNode{Type: "IF_NOT_EXISTS", Value: "true"})
	}
	root.Children = append(root.Children, *tableNode)
	i++

	// Debe haber paréntesis de apertura
	if i >= len(tokens) || tokens[i].Value != "(" {
		return nil, fmt.Errorf("se esperaba '(' después del nombre de tabla '%s'", tableNode.Value)
	}
	i++

	columnsNode := &SyntaxNode{Type: "COLUMNS"}
	columnCount := 0

	// Verificar que no esté vacío
	if i < len(tokens) && tokens[i].Value == ")" {
		return nil, fmt.Errorf("la definición de tabla no puede estar vacía")
	}

	for i < len(tokens) && tokens[i].Value != ")" {
		// Nombre de columna
		if tokens[i].Type != "IDENTIFICADOR" {
			return nil, fmt.Errorf("se esperaba nombre de columna, se encontró '%s'", tokens[i].Value)
		}

		columnName := tokens[i].Value
		columnDef := &SyntaxNode{Type: "COLUMN_DEFINITION", Value: columnName}
		i++

		// Tipo de dato
		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba tipo de dato para la columna '%s'", columnName)
		}

		// Mapa de tipos de datos válidos
		validTypes := map[string]bool{
			"INT": true, "INTEGER": true, "BIGINT": true, "SMALLINT": true,
			"SERIAL": true, "BIGSERIAL": true,
			"VARCHAR": true, "TEXT": true, "CHAR": true,
			"DECIMAL": true, "NUMERIC": true, "FLOAT": true, "REAL": true,
			"DOUBLE": true, "MONEY": true,
			"DATE": true, "TIME": true, "TIMESTAMP": true, "INTERVAL": true,
			"BOOLEAN": true, "BOOL": true,
			"UUID": true, "JSON": true, "JSONB": true,
			"ARRAY": true, "BYTEA": true,
		}

		upperType := strings.ToUpper(tokens[i].Value)
		if !validTypes[upperType] {
			// Verificar si es un tipo con palabras múltiples
			if upperType == "DOUBLE" && i+1 < len(tokens) &&
				strings.ToUpper(tokens[i+1].Value) == "PRECISION" {
				upperType = "DOUBLE PRECISION"
				i++
			} else {
				return nil, fmt.Errorf("tipo de dato inválido: '%s' para columna '%s'", tokens[i].Value, columnName)
			}
		}

		dataType := &SyntaxNode{Type: "DATA_TYPE", Value: upperType}
		columnDef.Children = append(columnDef.Children, *dataType)
		i++

		// Verificar parámetros del tipo (ej: VARCHAR(50))
		if i < len(tokens) && tokens[i].Value == "(" {
			i++
			if i >= len(tokens) {
				return nil, fmt.Errorf("se esperaba tamaño después de '(' en tipo %s", upperType)
			}

			// Para tipos como DECIMAL(10,2)
			sizeParams := []string{}

			if tokens[i].Type == "NUMERO" {
				sizeParams = append(sizeParams, tokens[i].Value)
				i++

				// Verificar si hay segundo parámetro (para DECIMAL)
				if i < len(tokens) && tokens[i].Value == "," {
					i++
					if i < len(tokens) && tokens[i].Type == "NUMERO" {
						sizeParams = append(sizeParams, tokens[i].Value)
						i++
					} else {
						return nil, fmt.Errorf("se esperaba segundo parámetro numérico después de ',' en %s", upperType)
					}
				}
			} else {
				return nil, fmt.Errorf("se esperaba número para el tamaño de %s", upperType)
			}

			if i >= len(tokens) || tokens[i].Value != ")" {
				return nil, fmt.Errorf("se esperaba ')' para cerrar los parámetros de %s", upperType)
			}
			i++

			for _, param := range sizeParams {
				dataType.Children = append(dataType.Children,
					SyntaxNode{Type: "SIZE", Value: param})
			}
		} else if upperType == "VARCHAR" || upperType == "CHAR" {
			// VARCHAR y CHAR deberían tener tamaño
			return nil, fmt.Errorf("tipo %s requiere especificar tamaño, ejemplo: %s(50)", upperType, upperType)
		}

		// Constraints
		for i < len(tokens) && tokens[i].Value != "," && tokens[i].Value != ")" {
			upperConstraint := strings.ToUpper(tokens[i].Value)

			switch upperConstraint {
			case "NOT":
				if i+1 >= len(tokens) {
					return nil, fmt.Errorf("se esperaba NULL después de NOT")
				}
				if strings.ToUpper(tokens[i+1].Value) != "NULL" {
					return nil, fmt.Errorf("se esperaba NULL después de NOT, se encontró '%s'", tokens[i+1].Value)
				}
				columnDef.Children = append(columnDef.Children,
					SyntaxNode{Type: "CONSTRAINT", Value: "NOT NULL"})
				i += 2

			case "NULL":
				columnDef.Children = append(columnDef.Children,
					SyntaxNode{Type: "CONSTRAINT", Value: "NULL"})
				i++

			case "PRIMARY":
				if i+1 >= len(tokens) {
					return nil, fmt.Errorf("se esperaba KEY después de PRIMARY")
				}
				if strings.ToUpper(tokens[i+1].Value) != "KEY" {
					return nil, fmt.Errorf("se esperaba KEY después de PRIMARY, se encontró '%s'", tokens[i+1].Value)
				}
				columnDef.Children = append(columnDef.Children,
					SyntaxNode{Type: "CONSTRAINT", Value: "PRIMARY KEY"})
				i += 2

			case "UNIQUE":
				columnDef.Children = append(columnDef.Children,
					SyntaxNode{Type: "CONSTRAINT", Value: "UNIQUE"})
				i++

			case "DEFAULT":
				i++
				if i >= len(tokens) {
					return nil, fmt.Errorf("se esperaba valor después de DEFAULT")
				}

				defaultValue := tokens[i].Value
				// Verificar valores especiales de DEFAULT
				upperDefault := strings.ToUpper(defaultValue)
				if upperDefault == "CURRENT_TIMESTAMP" || upperDefault == "NOW()" ||
					tokens[i].Type == "NUMERO" || tokens[i].Type == "CADENA" ||
					upperDefault == "TRUE" || upperDefault == "FALSE" ||
					upperDefault == "NULL" {
					columnDef.Children = append(columnDef.Children,
						SyntaxNode{Type: "DEFAULT", Value: defaultValue})
					i++
				} else {
					return nil, fmt.Errorf("valor DEFAULT inválido: '%s'", defaultValue)
				}

			case "REFERENCES":
				i++
				if i >= len(tokens) {
					return nil, fmt.Errorf("se esperaba nombre de tabla después de REFERENCES")
				}
				if tokens[i].Type != "IDENTIFICADOR" {
					return nil, fmt.Errorf("nombre de tabla inválido después de REFERENCES: '%s'", tokens[i].Value)
				}

				refNode := &SyntaxNode{Type: "REFERENCES", Value: tokens[i].Value}
				i++

				// Columna referenciada (opcional pero recomendada)
				if i < len(tokens) && tokens[i].Value == "(" {
					i++
					if i >= len(tokens) {
						return nil, fmt.Errorf("se esperaba nombre de columna después de '(' en REFERENCES")
					}
					if tokens[i].Type != "IDENTIFICADOR" {
						return nil, fmt.Errorf("nombre de columna inválido en REFERENCES: '%s'", tokens[i].Value)
					}
					refNode.Children = append(refNode.Children,
						SyntaxNode{Type: "REF_COLUMN", Value: tokens[i].Value})
					i++
					if i >= len(tokens) || tokens[i].Value != ")" {
						return nil, fmt.Errorf("se esperaba ')' después de la columna en REFERENCES")
					}
					i++
				}
				columnDef.Children = append(columnDef.Children, *refNode)

			case "CHECK":
				// CHECK constraint
				i++
				if i >= len(tokens) || tokens[i].Value != "(" {
					return nil, fmt.Errorf("se esperaba '(' después de CHECK")
				}
				i++

				// Capturar el contenido del CHECK
				checkDepth := 1
				checkContent := []string{}

				for i < len(tokens) && checkDepth > 0 {
					if tokens[i].Value == "(" {
						checkDepth++
					} else if tokens[i].Value == ")" {
						checkDepth--
						if checkDepth == 0 {
							break
						}
					}
					checkContent = append(checkContent, tokens[i].Value)
					i++
				}

				if checkDepth != 0 {
					return nil, fmt.Errorf("paréntesis no balanceados en constraint CHECK")
				}

				// Ahora sí usamos checkStart para algo útil
				checkNode := &SyntaxNode{
					Type:  "CONSTRAINT",
					Value: "CHECK",
				}

				// Guardar el contenido del CHECK
				if len(checkContent) > 0 {
					checkCondition := strings.Join(checkContent, " ")
					checkNode.Children = append(checkNode.Children,
						SyntaxNode{Type: "CHECK_CONDITION", Value: checkCondition})
				}

				columnDef.Children = append(columnDef.Children, *checkNode)
				i++ // Saltar el ')' final

			default:
				return nil, fmt.Errorf("constraint no reconocido: '%s' en columna '%s'", tokens[i].Value, columnName)
			}
		}

		columnsNode.Children = append(columnsNode.Children, *columnDef)
		columnCount++

		// Verificar si hay más columnas o constraints de tabla
		if i < len(tokens) && tokens[i].Value == "," {
			i++

			// Podría ser otra columna o un constraint de tabla
			if i >= len(tokens) {
				return nil, fmt.Errorf("se esperaba definición después de ','")
			}

			// Verificar constraints de tabla (PRIMARY KEY, FOREIGN KEY, etc.)
			if i+1 < len(tokens) {
				upperNext := strings.ToUpper(tokens[i].Value)
				if upperNext == "PRIMARY" || upperNext == "FOREIGN" ||
					upperNext == "UNIQUE" || upperNext == "CHECK" ||
					upperNext == "CONSTRAINT" {
					// Es un constraint de tabla, analizarlo
					tableConstraint, newIndex, err := analyzeTableConstraint(tokens, i)
					if err != nil {
						return nil, err
					}
					columnsNode.Children = append(columnsNode.Children, *tableConstraint)
					i = newIndex

					// Puede haber más constraints
					for i < len(tokens) && tokens[i].Value == "," {
						i++
						if i >= len(tokens) || tokens[i].Value == ")" {
							break
						}
						tableConstraint, newIndex, err = analyzeTableConstraint(tokens, i)
						if err != nil {
							return nil, err
						}
						columnsNode.Children = append(columnsNode.Children, *tableConstraint)
						i = newIndex
					}
				}
			}
		}
	}

	if columnCount == 0 {
		return nil, fmt.Errorf("se debe definir al menos una columna en la tabla")
	}

	if i >= len(tokens) || tokens[i].Value != ")" {
		return nil, fmt.Errorf("se esperaba ')' para cerrar la definición de tabla, se encontró: '%s'",
			func() string {
				if i < len(tokens) {
					return tokens[i].Value
				}
				return "fin de query"
			}())
	}

	root.Children = append(root.Children, *columnsNode)
	i++

	// Verificar punto y coma opcional al final
	if i < len(tokens) && tokens[i].Value != ";" {
		// Si hay más tokens y no es punto y coma, es un error
		return nil, fmt.Errorf("se esperaba ';' al final de CREATE TABLE, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeTableConstraint(tokens []Token, startIndex int) (*SyntaxNode, int, error) {
	i := startIndex
	constraint := &SyntaxNode{Type: "TABLE_CONSTRAINT"}

	// CONSTRAINT nombre (opcional)
	if strings.ToUpper(tokens[i].Value) == "CONSTRAINT" {
		i++
		if i >= len(tokens) || tokens[i].Type != "IDENTIFICADOR" {
			return nil, i, fmt.Errorf("se esperaba nombre después de CONSTRAINT")
		}
		constraint.Value = tokens[i].Value
		i++
	}

	if i >= len(tokens) {
		return nil, i, fmt.Errorf("se esperaba tipo de constraint")
	}

	upperConstraint := strings.ToUpper(tokens[i].Value)

	switch upperConstraint {
	case "PRIMARY":
		if i+1 >= len(tokens) || strings.ToUpper(tokens[i+1].Value) != "KEY" {
			return nil, i, fmt.Errorf("se esperaba KEY después de PRIMARY")
		}
		i += 2

		if i >= len(tokens) || tokens[i].Value != "(" {
			return nil, i, fmt.Errorf("se esperaba '(' después de PRIMARY KEY")
		}
		i++

		// Lista de columnas
		pkNode := &SyntaxNode{Type: "PRIMARY_KEY"}
		for i < len(tokens) && tokens[i].Value != ")" {
			if tokens[i].Type == "IDENTIFICADOR" {
				pkNode.Children = append(pkNode.Children,
					SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
				i++
				if i < len(tokens) && tokens[i].Value == "," {
					i++
				}
			} else {
				return nil, i, fmt.Errorf("se esperaba nombre de columna en PRIMARY KEY")
			}
		}

		if i >= len(tokens) || tokens[i].Value != ")" {
			return nil, i, fmt.Errorf("se esperaba ')' para cerrar PRIMARY KEY")
		}
		i++

		constraint.Children = append(constraint.Children, *pkNode)

	case "FOREIGN":
		if i+1 >= len(tokens) || strings.ToUpper(tokens[i+1].Value) != "KEY" {
			return nil, i, fmt.Errorf("se esperaba KEY después de FOREIGN")
		}
		i += 2

		if i >= len(tokens) || tokens[i].Value != "(" {
			return nil, i, fmt.Errorf("se esperaba '(' después de FOREIGN KEY")
		}
		i++

		// Columna local
		if i >= len(tokens) || tokens[i].Type != "IDENTIFICADOR" {
			return nil, i, fmt.Errorf("se esperaba nombre de columna en FOREIGN KEY")
		}

		fkNode := &SyntaxNode{Type: "FOREIGN_KEY", Value: tokens[i].Value}
		i++

		if i >= len(tokens) || tokens[i].Value != ")" {
			return nil, i, fmt.Errorf("se esperaba ')' después de la columna en FOREIGN KEY")
		}
		i++

		if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "REFERENCES" {
			return nil, i, fmt.Errorf("se esperaba REFERENCES después de FOREIGN KEY")
		}
		i++

		// Tabla referenciada
		if i >= len(tokens) || tokens[i].Type != "IDENTIFICADOR" {
			return nil, i, fmt.Errorf("se esperaba nombre de tabla después de REFERENCES")
		}

		refNode := &SyntaxNode{Type: "REFERENCES", Value: tokens[i].Value}
		i++

		if i < len(tokens) && tokens[i].Value == "(" {
			i++
			if i >= len(tokens) || tokens[i].Type != "IDENTIFICADOR" {
				return nil, i, fmt.Errorf("se esperaba nombre de columna en REFERENCES")
			}
			refNode.Children = append(refNode.Children,
				SyntaxNode{Type: "REF_COLUMN", Value: tokens[i].Value})
			i++
			if i >= len(tokens) || tokens[i].Value != ")" {
				return nil, i, fmt.Errorf("se esperaba ')' para cerrar REFERENCES")
			}
			i++
		}

		fkNode.Children = append(fkNode.Children, *refNode)
		constraint.Children = append(constraint.Children, *fkNode)

	case "UNIQUE":
		if i+1 < len(tokens) && tokens[i+1].Value == "(" {
			i += 2
			uniqueNode := &SyntaxNode{Type: "UNIQUE"}

			for i < len(tokens) && tokens[i].Value != ")" {
				if tokens[i].Type == "IDENTIFICADOR" {
					uniqueNode.Children = append(uniqueNode.Children,
						SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
					i++
					if i < len(tokens) && tokens[i].Value == "," {
						i++
					}
				} else {
					return nil, i, fmt.Errorf("se esperaba nombre de columna en UNIQUE")
				}
			}

			if i >= len(tokens) || tokens[i].Value != ")" {
				return nil, i, fmt.Errorf("se esperaba ')' para cerrar UNIQUE")
			}
			i++

			constraint.Children = append(constraint.Children, *uniqueNode)
		} else {
			return nil, i, fmt.Errorf("se esperaba '(' después de UNIQUE")
		}

	default:
		return nil, i, fmt.Errorf("tipo de constraint de tabla no reconocido: '%s'", tokens[i].Value)
	}

	return constraint, i, nil
}

// Mantener las demás funciones del código anterior...
// analyzeSelect, analyzeInsert, analyzeUpdate, analyzeDelete, analyzeDrop, etc.
// (las mismas que en la versión anterior)

func analyzeSelect(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "SELECT_STATEMENT"}
	i := 0

	// Verificar SELECT
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "SELECT" {
		return nil, fmt.Errorf("se esperaba SELECT")
	}
	i++

	// Verificar que hay columnas después de SELECT
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaban columnas después de SELECT")
	}

	// Verificar DISTINCT (opcional)
	hasDistinct := false
	if strings.ToUpper(tokens[i].Value) == "DISTINCT" {
		hasDistinct = true
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaban columnas después de DISTINCT")
		}
	}

	// Analizar columnas
	columnsNode := &SyntaxNode{Type: "COLUMNS"}
	if hasDistinct {
		columnsNode.Type = "DISTINCT_COLUMNS"
	}

	columnCount := 0
	expectingColumn := true

	for i < len(tokens) && strings.ToUpper(tokens[i].Value) != "FROM" {
		if tokens[i].Value == "," {
			if !expectingColumn {
				expectingColumn = true
			} else {
				return nil, fmt.Errorf("se esperaba una columna antes de ','")
			}
		} else if expectingColumn {
			if tokens[i].Type == "IDENTIFICADOR" || tokens[i].Value == "*" {
				columnsNode.Children = append(columnsNode.Children,
					SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
				columnCount++
				expectingColumn = false
			} else if tokens[i].Type == "PALABRA_CLAVE" &&
				(strings.ToUpper(tokens[i].Value) == "COUNT" ||
					strings.ToUpper(tokens[i].Value) == "SUM" ||
					strings.ToUpper(tokens[i].Value) == "AVG" ||
					strings.ToUpper(tokens[i].Value) == "MAX" ||
					strings.ToUpper(tokens[i].Value) == "MIN") {
				// Función agregada
				funcNode := &SyntaxNode{Type: "FUNCTION", Value: tokens[i].Value}
				i++
				if i < len(tokens) && tokens[i].Value == "(" {
					i++
					// Contenido de la función
					for i < len(tokens) && tokens[i].Value != ")" {
						i++
					}
					if i >= len(tokens) {
						return nil, fmt.Errorf("se esperaba ')' para cerrar la función")
					}
				}
				columnsNode.Children = append(columnsNode.Children, *funcNode)
				columnCount++
				expectingColumn = false
			} else {
				return nil, fmt.Errorf("se esperaba un nombre de columna o '*', se encontró '%s'", tokens[i].Value)
			}
		}
		i++
	}

	if columnCount == 0 {
		return nil, fmt.Errorf("se debe especificar al menos una columna después de SELECT")
	}

	if expectingColumn {
		return nil, fmt.Errorf("se esperaba una columna después de ','")
	}

	root.Children = append(root.Children, *columnsNode)

	// FROM es obligatorio en SELECT
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "FROM" {
		return nil, fmt.Errorf("se esperaba FROM después de las columnas")
	}
	i++

	// Tabla después de FROM
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de FROM")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("se esperaba un nombre de tabla válido después de FROM, se encontró '%s'", tokens[i].Value)
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	root.Children = append(root.Children, *tableNode)
	i++

	// Analizar cláusulas opcionales
	for i < len(tokens) {
		upperValue := strings.ToUpper(tokens[i].Value)

		switch upperValue {
		case "WHERE":
			whereNode, newIndex, err := analyzeWhereClause(tokens, i)
			if err != nil {
				return nil, err
			}
			root.Children = append(root.Children, *whereNode)
			i = newIndex

		case "GROUP":
			if i+1 < len(tokens) && strings.ToUpper(tokens[i+1].Value) == "BY" {
				groupNode, newIndex, err := analyzeGroupByClause(tokens, i)
				if err != nil {
					return nil, err
				}
				root.Children = append(root.Children, *groupNode)
				i = newIndex
			} else {
				return nil, fmt.Errorf("se esperaba BY después de GROUP")
			}

		case "ORDER":
			if i+1 < len(tokens) && strings.ToUpper(tokens[i+1].Value) == "BY" {
				orderNode, newIndex, err := analyzeOrderByClause(tokens, i)
				if err != nil {
					return nil, err
				}
				root.Children = append(root.Children, *orderNode)
				i = newIndex
			} else {
				return nil, fmt.Errorf("se esperaba BY después de ORDER")
			}

		case "LIMIT":
			if i+1 < len(tokens) && tokens[i+1].Type == "NUMERO" {
				limitNode := &SyntaxNode{Type: "LIMIT", Value: tokens[i+1].Value}
				root.Children = append(root.Children, *limitNode)
				i += 2
			} else {
				return nil, fmt.Errorf("se esperaba un número después de LIMIT")
			}

		case ";":
			i++

		default:
			return nil, fmt.Errorf("cláusula no reconocida: '%s'", tokens[i].Value)
		}
	}

	return root, nil
}

func analyzeInsert(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "INSERT_STATEMENT"}
	i := 0

	// Verificar INSERT
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "INSERT" {
		return nil, fmt.Errorf("se esperaba INSERT")
	}
	i++

	// Verificar INTO
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "INTO" {
		return nil, fmt.Errorf("se esperaba INTO después de INSERT")
	}
	i++

	// Nombre de tabla
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de INTO")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de tabla inválido: '%s'", tokens[i].Value)
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	root.Children = append(root.Children, *tableNode)
	i++

	// Columnas (opcional)
	columnsNode := &SyntaxNode{Type: "COLUMNS"}
	if i < len(tokens) && tokens[i].Value == "(" {
		i++

		columnCount := 0
		expectingColumn := true

		for i < len(tokens) && tokens[i].Value != ")" {
			if tokens[i].Value == "," {
				if !expectingColumn {
					expectingColumn = true
				} else {
					return nil, fmt.Errorf("se esperaba un nombre de columna antes de ','")
				}
			} else if expectingColumn && tokens[i].Type == "IDENTIFICADOR" {
				columnsNode.Children = append(columnsNode.Children,
					SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
				columnCount++
				expectingColumn = false
			} else if expectingColumn {
				return nil, fmt.Errorf("se esperaba un nombre de columna, se encontró '%s'", tokens[i].Value)
			}
			i++
		}

		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba ')' para cerrar la lista de columnas")
		}

		if columnCount == 0 {
			return nil, fmt.Errorf("se debe especificar al menos una columna")
		}

		if expectingColumn {
			return nil, fmt.Errorf("se esperaba un nombre de columna después de ','")
		}

		root.Children = append(root.Children, *columnsNode)
		i++
	}

	// VALUES es obligatorio
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "VALUES" {
		return nil, fmt.Errorf("se esperaba VALUES")
	}
	i++

	// Valores
	if i >= len(tokens) || tokens[i].Value != "(" {
		return nil, fmt.Errorf("se esperaba '(' después de VALUES")
	}

	valuesNode := &SyntaxNode{Type: "VALUES"}

	// Puede haber múltiples conjuntos de valores
	for i < len(tokens) && tokens[i].Value == "(" {
		i++
		valueSet := &SyntaxNode{Type: "VALUE_SET"}
		valueCount := 0
		expectingValue := true

		for i < len(tokens) && tokens[i].Value != ")" {
			if tokens[i].Value == "," {
				if !expectingValue {
					expectingValue = true
				} else {
					return nil, fmt.Errorf("se esperaba un valor antes de ','")
				}
			} else if expectingValue {
				if tokens[i].Type == "CADENA" || tokens[i].Type == "NUMERO" ||
					tokens[i].Type == "IDENTIFICADOR" {
					valueSet.Children = append(valueSet.Children,
						SyntaxNode{Type: "VALUE", Value: tokens[i].Value})
					valueCount++
					expectingValue = false
				} else {
					return nil, fmt.Errorf("tipo de valor inválido: '%s'", tokens[i].Value)
				}
			}
			i++
		}

		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba ')' para cerrar los valores")
		}

		if valueCount == 0 {
			return nil, fmt.Errorf("se debe especificar al menos un valor")
		}

		if expectingValue {
			return nil, fmt.Errorf("se esperaba un valor después de ','")
		}

		valuesNode.Children = append(valuesNode.Children, *valueSet)
		i++

		// Verificar si hay más conjuntos de valores
		if i < len(tokens) && tokens[i].Value == "," {
			i++
		}
	}

	if len(valuesNode.Children) == 0 {
		return nil, fmt.Errorf("se debe especificar al menos un conjunto de valores")
	}

	root.Children = append(root.Children, *valuesNode)

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de INSERT, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeUpdate(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "UPDATE_STATEMENT"}
	i := 0

	// Verificar UPDATE
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "UPDATE" {
		return nil, fmt.Errorf("se esperaba UPDATE")
	}
	i++

	// Nombre de tabla
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de UPDATE")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de tabla inválido: '%s'", tokens[i].Value)
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	root.Children = append(root.Children, *tableNode)
	i++

	// SET es obligatorio
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "SET" {
		return nil, fmt.Errorf("se esperaba SET después del nombre de tabla")
	}
	i++

	// Asignaciones
	setNode := &SyntaxNode{Type: "SET_CLAUSE"}
	assignmentCount := 0

	for i < len(tokens) && strings.ToUpper(tokens[i].Value) != "WHERE" && tokens[i].Value != ";" {
		// Columna
		if tokens[i].Type != "IDENTIFICADOR" {
			return nil, fmt.Errorf("se esperaba nombre de columna, se encontró '%s'", tokens[i].Value)
		}

		columnName := tokens[i].Value
		i++

		// Operador =
		if i >= len(tokens) || tokens[i].Value != "=" {
			return nil, fmt.Errorf("se esperaba '=' después de '%s'", columnName)
		}
		i++

		// Valor
		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba un valor después de '='")
		}

		if tokens[i].Type != "CADENA" && tokens[i].Type != "NUMERO" &&
			tokens[i].Type != "IDENTIFICADOR" {
			return nil, fmt.Errorf("tipo de valor inválido para asignación")
		}

		assignment := &SyntaxNode{
			Type: "ASSIGNMENT",
			Children: []SyntaxNode{
				{Type: "COLUMN", Value: columnName},
				{Type: "VALUE", Value: tokens[i].Value},
			},
		}
		setNode.Children = append(setNode.Children, *assignment)
		assignmentCount++
		i++

		// Verificar si hay más asignaciones
		if i < len(tokens) && tokens[i].Value == "," {
			i++
		}
	}

	if assignmentCount == 0 {
		return nil, fmt.Errorf("se debe especificar al menos una asignación después de SET")
	}

	root.Children = append(root.Children, *setNode)

	// WHERE (opcional pero recomendado)
	if i < len(tokens) && strings.ToUpper(tokens[i].Value) == "WHERE" {
		whereNode, newIndex, err := analyzeWhereClause(tokens, i)
		if err != nil {
			return nil, err
		}
		root.Children = append(root.Children, *whereNode)
		i = newIndex
	}

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de UPDATE, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeDelete(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "DELETE_STATEMENT"}
	i := 0

	// Verificar DELETE
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "DELETE" {
		return nil, fmt.Errorf("se esperaba DELETE")
	}
	i++

	// FROM es obligatorio
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "FROM" {
		return nil, fmt.Errorf("se esperaba FROM después de DELETE")
	}
	i++

	// Nombre de tabla
	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de FROM")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de tabla inválido: '%s'", tokens[i].Value)
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	root.Children = append(root.Children, *tableNode)
	i++

	// WHERE (opcional pero muy recomendado)
	if i < len(tokens) && strings.ToUpper(tokens[i].Value) == "WHERE" {
		whereNode, newIndex, err := analyzeWhereClause(tokens, i)
		if err != nil {
			return nil, err
		}
		root.Children = append(root.Children, *whereNode)
		i = newIndex
	}

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de DELETE, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeCreate(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "CREATE_STATEMENT"}
	i := 0

	// Verificar CREATE
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "CREATE" {
		return nil, fmt.Errorf("se esperaba CREATE")
	}
	i++

	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba TABLE o DATABASE después de CREATE")
	}

	upperValue := strings.ToUpper(tokens[i].Value)

	switch upperValue {
	case "TABLE":
		return analyzeCreateTable(tokens, i, root)
	case "DATABASE":
		return analyzeCreateDatabase(tokens, i, root)
	case "INDEX":
		return analyzeCreateIndex(tokens, i, root)
	default:
		return nil, fmt.Errorf("se esperaba TABLE, DATABASE o INDEX después de CREATE, se encontró '%s'", tokens[i].Value)
	}
}

func analyzeCreateDatabase(tokens []Token, startIndex int, root *SyntaxNode) (*SyntaxNode, error) {
	i := startIndex + 1

	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de base de datos después de CREATE DATABASE")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de base de datos inválido: '%s'", tokens[i].Value)
	}

	dbNode := &SyntaxNode{Type: "DATABASE", Value: tokens[i].Value}
	root.Children = append(root.Children, *dbNode)
	i++

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de CREATE DATABASE, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeCreateIndex(tokens []Token, startIndex int, root *SyntaxNode) (*SyntaxNode, error) {
	i := startIndex + 1

	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba nombre de índice después de CREATE INDEX")
	}

	if tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("nombre de índice inválido: '%s'", tokens[i].Value)
	}

	indexNode := &SyntaxNode{Type: "INDEX", Value: tokens[i].Value}
	root.Children = append(root.Children, *indexNode)
	i++

	// ON tabla
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "ON" {
		return nil, fmt.Errorf("se esperaba ON después del nombre del índice")
	}
	i++

	if i >= len(tokens) || tokens[i].Type != "IDENTIFICADOR" {
		return nil, fmt.Errorf("se esperaba nombre de tabla después de ON")
	}

	tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
	indexNode.Children = append(indexNode.Children, *tableNode)
	i++

	// Columnas
	if i >= len(tokens) || tokens[i].Value != "(" {
		return nil, fmt.Errorf("se esperaba '(' después del nombre de tabla")
	}
	i++

	columnsNode := &SyntaxNode{Type: "COLUMNS"}
	for i < len(tokens) && tokens[i].Value != ")" {
		if tokens[i].Type == "IDENTIFICADOR" {
			columnsNode.Children = append(columnsNode.Children,
				SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
			i++
			if i < len(tokens) && tokens[i].Value == "," {
				i++
			}
		} else {
			return nil, fmt.Errorf("se esperaba nombre de columna en CREATE INDEX")
		}
	}

	if i >= len(tokens) || tokens[i].Value != ")" {
		return nil, fmt.Errorf("se esperaba ')' para cerrar las columnas del índice")
	}
	i++

	indexNode.Children = append(indexNode.Children, *columnsNode)

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de CREATE INDEX, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

func analyzeDrop(tokens []Token) (*SyntaxNode, error) {
	root := &SyntaxNode{Type: "DROP_STATEMENT"}
	i := 0

	// Verificar DROP
	if i >= len(tokens) || strings.ToUpper(tokens[i].Value) != "DROP" {
		return nil, fmt.Errorf("se esperaba DROP")
	}
	i++

	if i >= len(tokens) {
		return nil, fmt.Errorf("se esperaba TABLE o DATABASE después de DROP")
	}

	upperValue := strings.ToUpper(tokens[i].Value)

	switch upperValue {
	case "TABLE":
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba nombre de tabla después de DROP TABLE")
		}
		if tokens[i].Type != "IDENTIFICADOR" {
			return nil, fmt.Errorf("nombre de tabla inválido: '%s'", tokens[i].Value)
		}
		tableNode := &SyntaxNode{Type: "TABLE", Value: tokens[i].Value}
		root.Children = append(root.Children, *tableNode)
		i++

	case "DATABASE":
		i++
		if i >= len(tokens) {
			return nil, fmt.Errorf("se esperaba nombre de base de datos después de DROP DATABASE")
		}
		if tokens[i].Type != "IDENTIFICADOR" {
			return nil, fmt.Errorf("nombre de base de datos inválido: '%s'", tokens[i].Value)
		}
		dbNode := &SyntaxNode{Type: "DATABASE", Value: tokens[i].Value}
		root.Children = append(root.Children, *dbNode)
		i++

	default:
		return nil, fmt.Errorf("se esperaba TABLE o DATABASE después de DROP, se encontró '%s'", tokens[i].Value)
	}

	// Verificar punto y coma opcional
	if i < len(tokens) && tokens[i].Value != ";" {
		return nil, fmt.Errorf("se esperaba ';' al final de DROP, se encontró: '%s'", tokens[i].Value)
	}

	return root, nil
}

// Funciones auxiliares
func analyzeWhereClause(tokens []Token, startIndex int) (*SyntaxNode, int, error) {
	whereNode := &SyntaxNode{Type: "WHERE_CLAUSE"}
	i := startIndex + 1

	if i >= len(tokens) {
		return nil, i, fmt.Errorf("se esperaba condición después de WHERE")
	}

	// Análisis simplificado de condición
	conditionCount := 0
	parenDepth := 0

	for i < len(tokens) &&
		(parenDepth > 0 || (strings.ToUpper(tokens[i].Value) != "GROUP" &&
			strings.ToUpper(tokens[i].Value) != "ORDER" &&
			strings.ToUpper(tokens[i].Value) != "LIMIT" &&
			tokens[i].Value != ";")) {

		if tokens[i].Value == "(" {
			parenDepth++
		} else if tokens[i].Value == ")" {
			parenDepth--
			if parenDepth < 0 {
				return nil, i, fmt.Errorf("paréntesis ')' inesperado en WHERE")
			}
		}

		whereNode.Children = append(whereNode.Children,
			SyntaxNode{Type: "CONDITION_TOKEN", Value: tokens[i].Value})
		conditionCount++
		i++
	}

	if parenDepth != 0 {
		return nil, i, fmt.Errorf("paréntesis no balanceados en condición WHERE")
	}

	if conditionCount == 0 {
		return nil, i, fmt.Errorf("WHERE requiere al menos una condición")
	}

	return whereNode, i, nil
}

func analyzeGroupByClause(tokens []Token, startIndex int) (*SyntaxNode, int, error) {
	groupNode := &SyntaxNode{Type: "GROUP_BY_CLAUSE"}
	i := startIndex + 2 // Saltar GROUP BY

	if i >= len(tokens) {
		return nil, i, fmt.Errorf("se esperaba columna después de GROUP BY")
	}

	columnCount := 0
	expectingColumn := true

	for i < len(tokens) &&
		strings.ToUpper(tokens[i].Value) != "HAVING" &&
		strings.ToUpper(tokens[i].Value) != "ORDER" &&
		strings.ToUpper(tokens[i].Value) != "LIMIT" &&
		tokens[i].Value != ";" {

		if tokens[i].Value == "," {
			if !expectingColumn {
				expectingColumn = true
			} else {
				return nil, i, fmt.Errorf("se esperaba columna antes de ','")
			}
		} else if expectingColumn && tokens[i].Type == "IDENTIFICADOR" {
			groupNode.Children = append(groupNode.Children,
				SyntaxNode{Type: "COLUMN", Value: tokens[i].Value})
			columnCount++
			expectingColumn = false
		} else if expectingColumn {
			return nil, i, fmt.Errorf("se esperaba nombre de columna en GROUP BY")
		}
		i++
	}

	if columnCount == 0 {
		return nil, i, fmt.Errorf("GROUP BY requiere al menos una columna")
	}

	if expectingColumn {
		return nil, i, fmt.Errorf("se esperaba columna después de ','")
	}

	// HAVING (opcional)
	if i < len(tokens) && strings.ToUpper(tokens[i].Value) == "HAVING" {
		i++
		havingNode := &SyntaxNode{Type: "HAVING_CLAUSE"}

		if i >= len(tokens) {
			return nil, i, fmt.Errorf("se esperaba condición después de HAVING")
		}

		// Condición de HAVING
		conditionCount := 0
		for i < len(tokens) &&
			strings.ToUpper(tokens[i].Value) != "ORDER" &&
			strings.ToUpper(tokens[i].Value) != "LIMIT" &&
			tokens[i].Value != ";" {

			havingNode.Children = append(havingNode.Children,
				SyntaxNode{Type: "CONDITION_TOKEN", Value: tokens[i].Value})
			conditionCount++
			i++
		}

		if conditionCount == 0 {
			return nil, i, fmt.Errorf("HAVING requiere al menos una condición")
		}

		groupNode.Children = append(groupNode.Children, *havingNode)
	}

	return groupNode, i, nil
}

func analyzeOrderByClause(tokens []Token, startIndex int) (*SyntaxNode, int, error) {
	orderNode := &SyntaxNode{Type: "ORDER_BY_CLAUSE"}
	i := startIndex + 2 // Saltar ORDER BY

	if i >= len(tokens) {
		return nil, i, fmt.Errorf("se esperaba columna después de ORDER BY")
	}

	columnCount := 0
	expectingColumn := true

	for i < len(tokens) &&
		strings.ToUpper(tokens[i].Value) != "LIMIT" &&
		tokens[i].Value != ";" {

		if tokens[i].Value == "," {
			if !expectingColumn {
				expectingColumn = true
			} else {
				return nil, i, fmt.Errorf("se esperaba columna antes de ','")
			}
		} else if expectingColumn && tokens[i].Type == "IDENTIFICADOR" {
			orderItem := &SyntaxNode{Type: "ORDER_ITEM", Value: tokens[i].Value}
			i++

			// Dirección (opcional)
			if i < len(tokens) &&
				(strings.ToUpper(tokens[i].Value) == "ASC" ||
					strings.ToUpper(tokens[i].Value) == "DESC") {
				orderItem.Children = append(orderItem.Children,
					SyntaxNode{Type: "DIRECTION", Value: tokens[i].Value})
				i++
			}

			orderNode.Children = append(orderNode.Children, *orderItem)
			columnCount++
			expectingColumn = false
		} else if expectingColumn {
			return nil, i, fmt.Errorf("se esperaba nombre de columna en ORDER BY")
		} else {
			i++
		}
	}

	if columnCount == 0 {
		return nil, i, fmt.Errorf("ORDER BY requiere al menos una columna")
	}

	if expectingColumn {
		return nil, i, fmt.Errorf("se esperaba columna después de ','")
	}

	return orderNode, i, nil
}
