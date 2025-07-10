import React from 'react';

const SyntacticAnalyzer = ({ result }) => {
  if (!result) {
    return (
      <div className="empty-state">
        <p>Ingresa una consulta SQL y haz clic en "Analizar Sintáctico" para ver el árbol sintáctico.</p>
      </div>
    );
  }

  if (!result.valid) {
    return (
      <div className="error">
        <p>Error en el análisis sintáctico: {result.error}</p>
      </div>
    );
  }

  const renderSyntaxTree = (node, level = 0) => {
    if (!node) return null;

    return (
      <div key={`${node.type}-${level}`} className="tree-node" style={{ marginLeft: `${level * 30}px` }}>
        <span className="node-type">{node.type}</span>
        {node.value && <span className="node-value">{node.value}</span>}
        {node.children && node.children.map((child, index) => 
          renderSyntaxTree(child, level + 1)
        )}
      </div>
    );
  };

  return (
    <div>
      <h2>Análisis Sintáctico</h2>
      <div className="syntax-tree">
        {renderSyntaxTree(result.syntax)}
      </div>
      <div style={{ marginTop: '20px', padding: '15px', background: '#0d1117', borderRadius: '8px', border: '1px solid #3fb950' }}>
        <h3 style={{ color: '#3fb950', marginBottom: '10px' }}>Sintaxis Correcta</h3>
        <p style={{ color: '#c9d1d9' }}>La estructura sintáctica de la consulta SQL es válida.</p>
      </div>
    </div>
  );
};

export default SyntacticAnalyzer;