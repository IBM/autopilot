import React from 'react';
import PropTypes from 'prop-types';

const Button = ({ text, color, onClick }) => {
    return (
        <button
            style={{
                backgroundColor: color,
                color: 'white',
                padding: '10px 20px',
                border: 'none',
                borderRadius: '5px',
                cursor: 'pointer',
                marginLeft: '20px'
            }}
            onClick={onClick}
        >
            {text}
        </button>
    );
};

Button.propTypes = {
    text: PropTypes.string.isRequired,
    color: PropTypes.string,
    onClick: PropTypes.func.isRequired // Function to execute on click
};

Button.defaultProps = {
    color: '#007BFF'
};

export default Button;
