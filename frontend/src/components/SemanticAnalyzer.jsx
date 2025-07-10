import React from 'react';

const SemanticAnalyzer = ({ result }) => {
  if (!result) {
    return (
      <div className="empty-state">
        <p>Ingresa una consulta SQL y haz clic en "Analizar Semántico" para verificar tablas y columnas.</p>
      </div>
    );
  }

  if (!result.valid && result.error) {
    return (
      <div className="error">
        <p>Error en el análisis semántico: {result.error}</p>
      </div>
    );
  }

  const semanticInfo = result.semantic || result;

  return (
    <div>
      <h2>Análisis Semántico</h2>
      
      {semanticInfo.tables && semanticInfo.tables.length > 0 && (
        <div className="semantic-section">
          <h3>Tablas Referenciadas</h3>
          {semanticInfo.tables.map((table, index) => (
            <div key={index} className="info-card">
              <span>{table.name}</span>
              <span className={table.exists ? 'exists-true' : 'exists-false'}>
                {table.exists ? 'Existe' : 'No existe'}
              </span>
            </div>
          ))}
        </div>
      )}

      {semanticInfo.columns && semanticInfo.columns.length > 0 && (
        <div className="semantic-section">
          <h3>Columnas Referenciadas</h3>
          {semanticInfo.columns.map((column, index) => (
            <div key={index} className="info-card">
              <span>{column.column}</span>
              <span className={column.exists ? 'exists-true' : 'exists-false'}>
                {column.exists ? 'Válida' : 'No encontrada'}
              </span>
            </div>
          ))}
        </div>
      )}

      {semanticInfo.warnings && semanticInfo.warnings.length > 0 && (
        <div className="semantic-section">
          <h3>Advertencias</h3>
          <ul className="warning-list">
            {semanticInfo.warnings.map((warning, index) => (
              <li key={index}>{warning}</li>
            ))}
          </ul>
        </div>
      )}

      {semanticInfo.valid && (
        <div style={{ marginTop: '20px', padding: '15px', background: '#0d1117', borderRadius: '8px', border: '1px solid #3fb950' }}>
          <h3 style={{ color: '#3fb950', marginBottom: '10px' }}>Análisis Semántico Correcto</h3>
          <p style={{ color: '#c9d1d9' }}>Todas las referencias a tablas y columnas son válidas.</p>
        </div>
      )}
    </div>
  );
};

export default SemanticAnalyzer;