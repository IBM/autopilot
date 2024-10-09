import React, { useState } from 'react';

function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');

  const handleLogin = (e) => {
    e.preventDefault();

    if (username === '' || password === '') {
      setError('Please fill in both fields!');
    } else {
      setError('');
      console.log('Logging in with:', { username, password });
    }
  };

  return (
    <div style={styles.loginContainer}>
      <div style={styles.loginBox}>
        <h2>Login</h2>
        {error && <p style={styles.errorMessage}>{error}</p>}
        <form onSubmit={handleLogin}>
          <div style={styles.formGroup}>
            <label>Username or Email</label>
            <input
              type="text"
              placeholder="Enter your username or email"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              style={styles.inputField}
            />
          </div>
          <div style={styles.formGroup}>
            <label>Password</label>
            <input
              type="password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              style={styles.inputField}
            />
          </div>
          <button type="submit" style={styles.loginButton}>
            Login
          </button>
          <a href="https://openshift-login-url" style={styles.openshiftButton}>
            OpenShift Login
          </a>
        </form>
      </div>
    </div>
  );
}

const styles = {
  loginContainer: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100vh',
    marginLeft: '200px',  // Adjusting for the sidebar
    width: 'calc(100% - 350px)',  // Width accounting for the 200px sidebar
    backgroundColor: '#f4f4f4',
    fontFamily: 'Arial, sans-serif',  // Apply Arial font here
  },
  loginBox: {
    backgroundColor: 'white',
    padding: '20px',
    borderRadius: '8px',
    boxShadow: '0 4px 8px rgba(0, 0, 0, 0.1)',
    width: '400px',
    textAlign: 'center',
    fontFamily: 'Arial, sans-serif',  // Apply Arial font here
  },
  formGroup: {
    marginBottom: '15px',
    textAlign: 'left',
  },
  inputField: {
    width: '95%',
    padding: '10px',
    border: '1px solid #ccc',
    borderRadius: '4px',
    marginBottom: '10px',
  },
  loginButton: {
    width: '95%',
    padding: '10px',
    color: 'white',
    backgroundColor: '#28a745',  // Green color for the Login button
    border: 'none',
    borderRadius: '4px',
    cursor: 'pointer',
    fontSize: '16px',
    marginBottom: '10px',
  },
  openshiftButton: {
    width: '90%',
    padding: '10px',
    backgroundColor: '#007bff',  // Blue color for the OpenShift button
    color: 'white',
    borderRadius: '4px',
    textAlign: 'center',
    textDecoration: 'none',
    display: 'inline-block',
    fontSize: '16px',
    marginBottom: '10px',
    cursor: 'pointer',
  },
  errorMessage: {
    color: 'red',
    marginBottom: '10px',
  },
};

export default Login;
