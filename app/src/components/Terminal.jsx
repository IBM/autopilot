import React, { useState } from 'react';
import { CopyButton } from '@carbon/react';

const Terminal = ({ output }) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(output || "No tests deployed yet...");
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div style={styles.container}>
            <div style={styles.terminal}>
                <pre style={styles.output}>{output || "No tests deployed yet..."}</pre>
            </div>
            <div style={styles.copyButton}>
                <CopyButton
                    feedback={copied ? "Copied!" : "Copy"}
                    onClick={handleCopy}
                    feedbackTimeout={1000}
                />
            </div>
        </div>
    );
}

const styles = {
    container: {
        position: 'relative',
    },
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
    },
    copyButton: {
        position: 'absolute',
        top: '10px',
        right: '10px',
    },
};

export default Terminal;