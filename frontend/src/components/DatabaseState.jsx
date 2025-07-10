import React from 'react';

const DatabaseState = ({ state }) => {
  if (!state || !state.tables) return null;

  return (
    <div className="database-state">
      <h3>Estado Actual de la Base de Datos</h3>
      <div className="db-summary">
        <div className="summary-card">
          <h4>Total de Tablas</h4>
          <p className="big-number">{state.totalTables}</p>
        </div>
      </div>
      
      <div className="tables-grid">
        {state.tables.map((table, index) => (
          <div key={index} className="table-card">
            <div className="table-header">
              <h4>{table.name}</h4>
            </div>
            <div className="table-info">
              <p>Registros: <strong>{table.rowCount}</strong></p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default DatabaseState;