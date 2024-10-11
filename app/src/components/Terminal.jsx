import React from 'react';

const Terminal = ({ output }) => {
    return (
        <div style={styles.terminal}>
            <pre style={styles.output}>{output || "No tests deployed yet..."}</pre>
        </div>
    );
}

const styles = {
    terminal: {
        backgroundColor: '#1e1e1e',
        color: 'white',
        padding: '10px',
        borderRadius: '5px',
        height: '50vh',
        overflowY: 'auto',
        fontFamily: 'monospace',
        whiteSpace: 'pre-wrap',
    },
    output: {
        margin: 0,
        lineHeight: '1.5',
    }
};

export default Terminal;
