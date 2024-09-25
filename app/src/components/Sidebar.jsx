import React from 'react';
import { Link } from 'react-router-dom';

function Sidebar() {
    return (
        <div style={styles.sidebar}>
            <h2>Navigation</h2>
            <ul style={styles.list}>
                <li>
                    <Link to="/login" style={styles.link}>Login</Link>
                </li>
                <li>
                    <Link to="/monitor" style={styles.link}>Monitor Cluster</Link>
                </li>
                <li>
                    <Link to="/testing" style={styles.link}>Run Tests</Link>
                </li>
                <li>
                    <Link to="/login" style={styles.link}>Log Out</Link>
                </li>
            </ul>
        </div>
    );
}

const styles = {
    sidebar: {
        width: '200px',
        height: '100vh',
        padding: '20px',
        backgroundColor: '#f4f4f4',
        position: 'fixed',
        top: 0,
        left: 0,
    },
    list: {
        listStyleType: 'none',
        padding: 0,
    },
    link: {
        textDecoration: 'none',
        color: '#333',
        margin: '10px 0',
        display: 'block',
    }
};

export default Sidebar;
