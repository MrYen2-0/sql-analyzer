import React, { useState, useEffect } from 'react';
import LexicalAnalyzer from './components/LexicalAnalyzer';
import SyntacticAnalyzer from './components/SyntacticAnalyzer';
import SemanticAnalyzer from './components/SemanticAnalyzer';
import QueryResults from './components/QueryResults';
import DatabaseState from './components/DatabaseState';
import axios from 'axios';

const API_URL = 'http://localhost:8080/api';

function App() {
  const [query, setQuery] = useState('');
  const [activeTab, setActiveTab] = useState('lexical');
  const [lexicalResult, setLexicalResult] = useState(null);
  const [syntacticResult, setSyntacticResult] = useState(null);
  const [semanticResult, setSemanticResult] = useState(null);
  const [queryResult, setQueryResult] = useState(null);
  const [databaseState, setDatabaseState] = useState(null);
  const [error, setError] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [loading, setLoading] = useState(false);

  // Cargar estado inicial de la base de datos
  useEffect(() => {
    loadDatabaseState();
  }, []);

  const loadDatabaseState = async () => {
    try {
      const response = await axios.get(`${API_URL}/database/state`);
      if (response.data.success) {
        setDatabaseState(response.data.state);
      }
    } catch (err) {
      console.error('Error cargando estado de BD:', err);
    }
  };

  const analyzeLexical = async () => {
    try {
      setError('');
      setLoading(true);
      const response = await axios.post(`${API_URL}/analyze/lexical`, { query });
      setLexicalResult(response.data);
      setActiveTab('lexical');
    } catch (err) {
      setError('Error en análisis léxico: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const analyzeSyntactic = async () => {
    try {
      setError('');
      setLoading(true);
      const response = await axios.post(`${API_URL}/analyze/syntactic`, { query });
      setSyntacticResult(response.data);
      setActiveTab('syntactic');
    } catch (err) {
      setError('Error en análisis sintáctico: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const analyzeSemantic = async () => {
    try {
      setError('');
      setLoading(true);
      const response = await axios.post(`${API_URL}/analyze/semantic`, { query });
      setSemanticResult(response.data);
      setActiveTab('semantic');
    } catch (err) {
      setError('Error en análisis semántico: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const executeQuery = async () => {
    try {
      setError('');
      setSuccessMessage('');
      setLoading(true);
      const response = await axios.post(`${API_URL}/execute`, { query });
      
      if (response.data.success) {
        setSuccessMessage('Query ejecutada exitosamente');
        setQueryResult(response.data.result);
        setDatabaseState(response.data.dbState);
        setActiveTab('results');
      } else {
        setError(response.data.error);
      }
    } catch (err) {
      setError('Error ejecutando query: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  // Ejemplos de queries
  const exampleQueries = [
    { label: "SELECT simple", query: "SELECT * FROM usuarios;" },
    { label: "INSERT", query: "INSERT INTO usuarios (nombre, email) VALUES ('Ana López', 'ana@email.com');" },
    { label: "UPDATE", query: "UPDATE productos SET precio = 999.99 WHERE nombre = 'Laptop Dell';" },
    { label: "DELETE", query: "DELETE FROM usuarios WHERE email = 'ana@email.com';" },
    { label: "CREATE TABLE", query: "CREATE TABLE categorias (id SERIAL PRIMARY KEY, nombre VARCHAR(50) NOT NULL);" },
  ];

  return (
    <div className="container">
      <header>
        <h1>Analizador Léxico, Sintáctico y Semántico</h1>
        <p className="subtitle">Con visualización de cambios en PostgreSQL</p>
      </header>

      <div className="main-content">
        <div className="left-panel">
          <div className="input-section">
            <textarea
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Escribe tu consulta SQL aquí..."
              rows="6"
              disabled={loading}
            />
            
            <div className="button-group">
              <button onClick={analyzeLexical} disabled={loading || !query}>
                Analizar Léxico
              </button>
              <button onClick={analyzeSyntactic} disabled={loading || !query}>
                Analizar Sintáctico
              </button>
              <button onClick={analyzeSemantic} disabled={loading || !query}>
                Analizar Semántico
              </button>
              <button onClick={executeQuery} className="execute-btn" disabled={loading || !query}>
                {loading ? 'Ejecutando...' : 'Ejecutar en PostgreSQL'}
              </button>
            </div>

            <div className="examples">
              <h4>Ejemplos rápidos:</h4>
              <div className="example-buttons">
                {exampleQueries.map((example, index) => (
                  <button
                    key={index}
                    className="example-btn"
                    onClick={() => setQuery(example.query)}
                  >
                    {example.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {error && <div className="error">{error}</div>}
          {successMessage && <div className="success">{successMessage}</div>}

          <div className="tabs">
            <button 
              className={activeTab === 'lexical' ? 'active' : ''}
              onClick={() => setActiveTab('lexical')}
            >
              Análisis Léxico
            </button>
            <button 
              className={activeTab === 'syntactic' ? 'active' : ''}
              onClick={() => setActiveTab('syntactic')}
            >
              Análisis Sintáctico
            </button>
            <button 
              className={activeTab === 'semantic' ? 'active' : ''}
              onClick={() => setActiveTab('semantic')}
            >
              Análisis Semántico
            </button>
            <button 
              className={activeTab === 'results' ? 'active' : ''}
              onClick={() => setActiveTab('results')}
            >
              Resultados
            </button>
          </div>

          <div className="tab-content">
            {activeTab === 'lexical' && <LexicalAnalyzer result={lexicalResult} />}
            {activeTab === 'syntactic' && <SyntacticAnalyzer result={syntacticResult} />}
            {activeTab === 'semantic' && <SemanticAnalyzer result={semanticResult} />}
            {activeTab === 'results' && <QueryResults result={queryResult} dbState={databaseState} />}
          </div>
        </div>

        <div className="right-panel">
          <DatabaseState state={databaseState} />
        </div>
      </div>
    </div>
  );
}

export default App;