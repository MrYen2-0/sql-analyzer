package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sql-analyzer/analyzer"
	"sql-analyzer/database"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type AnalyzeRequest struct {
	Query string `json:"query"`
}

type AnalyzeResponse struct {
	Valid    bool             `json:"valid"`
	Tokens   []analyzer.Token `json:"tokens,omitempty"`
	Syntax   interface{}      `json:"syntax,omitempty"`
	Semantic interface{}      `json:"semantic,omitempty"`
	Error    string           `json:"error,omitempty"`
}

func main() {
	r := mux.NewRouter()

	// Rutas
	r.HandleFunc("/api/analyze/lexical", handleLexicalAnalysis).Methods("POST")
	r.HandleFunc("/api/analyze/syntactic", handleSyntacticAnalysis).Methods("POST")
	r.HandleFunc("/api/analyze/semantic", handleSemanticAnalysis).Methods("POST")
	r.HandleFunc("/api/execute", handleExecuteQuery).Methods("POST")

	// Nueva ruta para obtener el estado de la base de datos
	r.HandleFunc("/api/database/state", handleDatabaseState).Methods("GET")

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
	})

	handler := c.Handler(r)
	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func handleLexicalAnalysis(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokens, err := analyzer.LexicalAnalysis(req.Query)
	if err != nil {
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(AnalyzeResponse{
		Valid:  true,
		Tokens: tokens,
	})
}

func handleSyntacticAnalysis(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	syntaxTree, err := analyzer.SyntacticAnalysis(req.Query)
	if err != nil {
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(AnalyzeResponse{
		Valid:  true,
		Syntax: syntaxTree,
	})
}

func handleSemanticAnalysis(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	semanticInfo, err := analyzer.SemanticAnalysis(req.Query)
	if err != nil {
		json.NewEncoder(w).Encode(AnalyzeResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(AnalyzeResponse{
		Valid:    true,
		Semantic: semanticInfo,
	})
}

func handleExecuteQuery(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validar con los tres análisis
	_, lexErr := analyzer.LexicalAnalysis(req.Query)
	if lexErr != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error léxico: " + lexErr.Error(),
		})
		return
	}

	_, synErr := analyzer.SyntacticAnalysis(req.Query)
	if synErr != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error sintáctico: " + synErr.Error(),
		})
		return
	}

	_, semErr := analyzer.SemanticAnalysis(req.Query)
	if semErr != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error semántico: " + semErr.Error(),
		})
		return
	}

	// Ejecutar en PostgreSQL
	result, err := database.ExecuteQuery(req.Query)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Obtener el estado actualizado de la base de datos
	dbState, _ := database.GetDatabaseState()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"result":  result,
		"dbState": dbState,
	})
}

func handleDatabaseState(w http.ResponseWriter, r *http.Request) {
	state, err := database.GetDatabaseState()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"state":   state,
	})
}
