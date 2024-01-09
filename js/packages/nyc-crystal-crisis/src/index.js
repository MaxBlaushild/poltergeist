import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import reportWebVitals from './reportWebVitals';
import { createTheme, ThemeProvider } from '@mui/material/styles';
import { Toaster } from 'react-hot-toast';

const theme = createTheme({
  overrides: {
    // Style the input
    MuiInputBase: {
      input: {
        fontFamily: 'Poppins', // Change to your desired font family
      },
    },
    // Style the label
    MuiInputLabel: {
      root: {
        fontFamily: 'Poppins', // Change to your desired font family
      },
    },
    // Style the placeholder
    MuiInputBase: {
      input: {
        '&::placeholder': {
          fontFamily: 'Poppins', // Change to your desired font family
        },
      },
    },
    MuiTypography: {
      fontFamily: 'Poppins',
    },
  },
});

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      {/* <Toaster> */}
      <App />
      {/* </Toaster> */}
    </ThemeProvider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
