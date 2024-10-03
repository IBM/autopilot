import React from 'react';

const Switch = ({ isOn, handleToggle, onColor = '#06D6A0', offColor = '#ccc' }) => {
  const switchStyles = {
    width: '50px',
    height: '25px',
    borderRadius: '25px',
    backgroundColor: isOn ? onColor : offColor,
    position: 'relative',
    cursor: 'pointer',
    transition: 'background-color 0.3s',
  };

  const buttonStyles = {
    width: '23px',
    height: '23px',
    backgroundColor: 'white',
    borderRadius: '50%',
    position: 'absolute',
    top: '1px',
    left: '1px',
    transform: isOn ? 'translateX(26px)' : 'translateX(0)',
    transition: 'transform 0.3s',
  };

  return (
    <div style={switchStyles} onClick={handleToggle}>
      <div style={buttonStyles} />
    </div>
  );
};

export default Switch;