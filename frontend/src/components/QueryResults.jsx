import React from 'react';

const QueryResults = ({ result, dbState }) => {
  if (!result) return null;

  const renderResultTable = () => {
    if (!result.data || result.data.length === 0) {
      return null;
    }

    return (
      <div className="result-table-container">
        <h3>Datos {result.type === 'DELETE' ? 'Eliminados' : 'Resultantes'}:</h3>
        <div className="table-wrapper">
          <table className="result-table">
            <thead>
              <tr>
                {result.columns.map((col, index) => (
                  <th key={index}>{col}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {result.data.map((row, rowIndex) => (
                <tr key={rowIndex}>
                  {result.columns.map((col, colIndex) => (
                    <td key={colIndex}>
                      {row[col] !== null ? String(row[col]) : 'NULL'}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    );
  };

  const getResultColor = () => {
    switch (result.type) {
      case 'INSERT':
      case 'CREATE':
        return '#3fb950';
      case 'UPDATE':
        return '#58a6ff';
      case 'DELETE':
      case 'DROP':
        return '#f85149';
      case 'SELECT':
        return '#a371f7';
      default:
        return '#8b949e';
    }
  };

  return (
    <div className="query-results">
      <div className="result-header" style={{ borderLeftColor: getResultColor() }}>
        <div className="result-info">
          <h3>{result.type} Ejecutado</h3>
          <p className="result-message">{result.message}</p>
          {result.rowsAffected !== undefined && result.rowsAffected > 0 && (
            <p className="rows-affected">
              Filas afectadas: <strong>{result.rowsAffected}</strong>
            </p>
          )}
        </div>
      </div>

      {renderResultTable()}

      {result.type === 'CREATE' && result.tableName && (
        <div className="schema-info">
          <h4>Estructura de la tabla '{result.tableName}':</h4>
          <table className="schema-table">
            <thead>
              <tr>
                <th>Columna</th>
                <th>Tipo</th>
                <th>Nullable</th>
                <th>Default</th>
              </tr>
            </thead>
            <tbody>
              {result.data && result.data.map((col, index) => (
                <tr key={index}>
                  <td>{col.column_name}</td>
                  <td>{col.data_type}</td>
                  <td>{col.is_nullable}</td>
                  <td>{col.column_default || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default QueryResults;