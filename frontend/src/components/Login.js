import React, { useState } from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';
import { Box, TextField, Button, Typography, Paper } from '@mui/material';

const Login = () => {
  const [inputUsername, setInputUsername] = useState('');
  const { setUsername, isConnected } = useWebSocket();

  const handleSubmit = (e) => {
    e.preventDefault();
    if (inputUsername.trim()) {
      setUsername(inputUsername.trim());
    }
  };

  return (
    <Box
      sx={{
        height: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        bgcolor: 'background.default'
      }}
    >
      <Paper
        elevation={3}
        sx={{
          p: 4,
          width: '100%',
          maxWidth: 400,
          textAlign: 'center'
        }}
      >
        <Typography variant="h4" component="h1" gutterBottom>
          ChatSync
        </Typography>
        <form onSubmit={handleSubmit}>
          <TextField
            fullWidth
            label="Username"
            value={inputUsername}
            onChange={(e) => setInputUsername(e.target.value)}
            margin="normal"
            required
          />
          <Button
            type="submit"
            variant="contained"
            fullWidth
            sx={{ mt: 2 }}
            disabled={!inputUsername.trim() || isConnected}
          >
            {isConnected ? 'Connected' : 'Connect'}
          </Button>
        </form>
      </Paper>
    </Box>
  );
};

export default Login; 