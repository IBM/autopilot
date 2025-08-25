import React, { useState } from 'react';
import PropTypes from 'prop-types';

const Dropdown = ({ title, options, placeholder, onSelect }) => {
    const [selectedOptions, setSelectedOptions] = useState('');

    const handleSelect = (event) => {
        const value = event.target.value;
        setSelectedOptions(value);
        onSelect(value);
    };

    return (
        <div>
            <label>
                <strong>{title}</strong>
            </label>
            <select
                value={selectedOptions}
                onChange={handleSelect}
                style={{
                    display: 'block',
                    width: '100%',
                    padding: '10px',
                    margin: '10px 0',
                    fontSize: '16px',
                    borderRadius: '5px',
                    border: '1px solid #ccc',
                }}
            >
                <option value="" disabled>
                    {placeholder}
                </option>
                {options.map((option, index) => (
                    <option key={index} value={option}>
                        {option}
                    </option>
                ))}
            </select>
        </div>
    );
};

Dropdown.propTypes = {
    options: PropTypes.array.isRequired,
    onSelect: PropTypes.func.isRequired,
    placeholder: PropTypes.string,
    title: PropTypes.string,
};

Dropdown.defaultProps = {
    placeholder: 'Select an option',
    title: '',
};

export default Dropdown;
