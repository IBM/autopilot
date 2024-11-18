import React, { useState } from 'react';
import {
  TextInput,
  PasswordInput,
  Button,
  InlineNotification,
  Grid,
  Column,
} from '@carbon/react';
import '@carbon/styles/css/styles.css'; // Correct Carbon styles import

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
    <Grid
      fullWidth
      style={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#ffffff',
        padding: '1rem',
      }}
    >
      <Column
        lg={4} md={6} sm={12}
        style={{
          backgroundColor: '#ffffff',
          padding: '2rem',
          borderRadius: '8px',
          boxShadow: '0 4px 8px rgba(0, 0, 0, 0.1)',
          maxWidth: '400px',
          width: '100%',
        }}
      >
        <form onSubmit={handleLogin}>
          <h2 style={{ textAlign: 'center', marginBottom: '1rem' }}>Login</h2>
          {error && (
            <InlineNotification
              kind="error"
              title="Error"
              subtitle={error}
              lowContrast
              style={{ marginBottom: '1rem' }}
            />
          )}
          <div style={{ marginBottom: '1rem' }}>
            <TextInput
              id="username"
              labelText="Username or Email"
              placeholder="Enter your username or email"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
            />
          </div>
          <div style={{ marginBottom: '1rem' }}>
            <PasswordInput
              id="password"
              labelText="Password"
              placeholder="Enter your password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
            />
          </div>

          {/* Centralized Button Container */}
          <div style={{ textAlign: 'center' }}>
            <Button
              type="submit"
              kind="primary"
              style={{
                marginBottom: '1rem',
                minWidth: '336px',
                height: '48px',
              }}
            >
              Login
            </Button>
            <Button
              kind="secondary"
              href="https://openshift-login-url"
              style={{
                minWidth: '336px',
                height: '48px',
              }}
            >
              OpenShift Login
            </Button>
          </div>
        </form>
      </Column>
    </Grid>
  );
}

export default Login;
