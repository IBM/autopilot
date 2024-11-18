import React from 'react';

const NumberField = ({ isDisabled, placeholder = "Enter a number...", value, onChange, min, max }) => {
  return (
    <input
      type="number"
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      disabled={isDisabled}
      min={min}
      max={max}
      
      style={{
        padding: '10px',
        fontSize: '16px',
        border: '1px solid #ccc',
        borderRadius: '4px',
        width: '150px',
        backgroundColor: isDisabled ? '#f0f0f0' : 'white',
        cursor: isDisabled ? 'not-allowed' : 'text',
      }}
    />
  );
};

export default NumberField;