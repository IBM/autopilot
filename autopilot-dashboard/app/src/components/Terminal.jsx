import React, { useState } from 'react';
import { Button, CopyButton, Modal } from '@carbon/react';
import { ExpandAll, Close } from '@carbon/icons-react';

const Terminal = ({ output }) => {
    const [copied, setCopied] = useState(false);
    const [isModalOpen, setModalOpen] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(output || "No tests deployed yet...");
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleOpenModal = () => {
        setModalOpen(true);
    };

    const handleCloseModal = () => {
        setModalOpen(false);
    };

    return (
        <div style={styles.container}>
            <div style={styles.terminal}>
                <pre style={styles.output}>{output || "No tests deployed yet..."}</pre>
            </div>
            <div style={styles.iconsContainer}>
                <CopyButton
                    feedback={copied ? "Copied!" : "Copy"}
                    onClick={handleCopy}
                    feedbackTimeout={1000}
                />
                <Button
                    renderIcon={ExpandAll}
                    hasIconOnly
                    onClick={handleOpenModal}
                    style={styles.expandIcon}
                    size="md"
                    iconDescription="Expand"
                    tooltipPosition="bottom"
                />
            </div>

            {isModalOpen && (
                <Modal
                    open={isModalOpen}
                    onRequestClose={handleCloseModal}
                    modalHeading="Test Results"
                    passiveModal
                    size="lg"
                >
                    <div style={styles.container}>
                        <div style={styles.modalTerminal}>
                            <pre style={styles.output}>{output || "No tests deployed yet..."}</pre>
                        </div>
                        <div style={styles.iconsContainer}>
                            <CopyButton
                                feedback={copied ? "Copied!" : "Copy"}
                                onClick={handleCopy}
                                feedbackTimeout={1000}
                            />
                        </div>
                    </div>
                </Modal>
            )}
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
    iconsContainer: {
        position: 'absolute',
        top: '10px',
        right: '10px',
        display: 'flex',
        gap: '10px',
    },
    expandIcon: {
        cursor: 'pointer',
    },
    modalTerminal: {
        backgroundColor: '#1e1e1e',
        color: 'white',
        padding: '20px',
        borderRadius: '5px',
        height: '80vh',
        overflowY: 'auto',
        fontFamily: 'monospace',
        whiteSpace: 'pre-wrap',
    },
};

export default Terminal;