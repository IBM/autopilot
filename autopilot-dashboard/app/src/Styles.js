export const containerStyle = {
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    padding: '20px',
};

export const sectionStyle = {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    gap: '20px',
    padding: '20px',
    backgroundColor: '#f4f4f4',
    boxShadow: '0px 4px 8px rgba(0, 0, 0, 0.1)',
    margin: '0 auto',
};

export const headerStyle = {
    color: '#3D3D3D',
    margin: '2vh 0',
    padding: '1vh',
    borderBottom: '0.2vh solid #E0E0E0',
    textAlign: 'center',
};

export const buttonStyle = {
    alignSelf: 'center',
    width: '10vw',
    paddingRight: '0vw',
};

export const dynamicWidth = (maxLength) => ({
    width: `${Math.max(200, Math.min(400, maxLength * 12))}px`,
});

// Add other styles in the same manner, for example:
export const testParameterStyle = {
    display: 'flex',
    gap: '2.5vw',
    justifyContent: 'center',
};

