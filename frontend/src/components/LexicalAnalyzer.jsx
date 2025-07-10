import React from 'react';

const LexicalAnalyzer = ({ result }) => {
  if (!result) {
    return (
      <div className="empty-state">
        <p>Ingresa una consulta SQL y haz clic en "Analizar Léxico" para ver los tokens.</p>
      </div>
    );
  }

  if (!result.valid) {
    return (
      <div className="error">
        <p>Error en el análisis léxico: {result.error}</p>
      </div>
    );
  }

  // Contar tokens por tipo
  const tokenCounts = result.tokens.reduce((acc, token) => {
    acc[token.tipo] = (acc[token.tipo] || 0) + 1;
    return acc;
  }, {});

  return (
    <div>
      <h2>Resultados del Análisis Léxico</h2>
      
      <table>
        <thead>
          <tr>
            <th>Token</th>
            <th>Tipo</th>
          </tr>
        </thead>
        <tbody>
          {result.tokens.map((token, index) => (
            <tr key={index}>
              <td>{token.token}</td>
              <td>{token.tipo}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="token-count">
        <h3 style={{gridColumn: '1/-1', textAlign: 'center'}}>Conteo de Tokens</h3>
        {Object.entries(tokenCounts).map(([type, count]) => (
          <div key={type} className="count-item">
            <h4>{type}</h4>
            <p>{count}</p>
          </div>
        ))}
        <div className="count-item">
          <h4>TOTAL</h4>
          <p>{result.tokens.length}</p>
        </div>
      </div>
    </div>
  );
};

export default LexicalAnalyzer;