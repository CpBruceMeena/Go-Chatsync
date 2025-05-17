import React from 'react';
import { ThemeProvider, CssBaseline } from '@mui/material';
import { WebSocketProvider, useWebSocket } from './contexts/WebSocketContext';
import Login from './components/Login';
import Chat from './components/Chat';
import theme from './theme';

const AppContent = () => {
  const { isConnected } = useWebSocket();
  return isConnected ? <Chat /> : <Login />;
};

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <WebSocketProvider>
        <AppContent />
      </WebSocketProvider>
    </ThemeProvider>
  );
}

export default App;
