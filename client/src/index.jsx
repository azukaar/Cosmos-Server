import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import customParseFormat from 'dayjs/plugin/customParseFormat'; // import this if you need to parse custom formats
import { AdapterDayjs } from '@mui/x-date-pickers/AdapterDayjs';
import localizedFormat from 'dayjs/plugin/localizedFormat'; // import this for localized formatting
import 'dayjs/locale/en-gb';
import 'dayjs/locale/en';
import 'dayjs/locale/de';

// import i18n (needs to be bundled ;)) 
import './utils/locales/i18n';
import { i18n } from './utils/locales/i18n';

// scroll bar
import 'simplebar/src/simplebar.css';

// third-party
import { Provider as ReduxProvider } from 'react-redux';

// apex-chart
import './assets/third-party/apex-chart.css';

import './index.css';

// project import
import App from './App';
import { store } from './store';
import reportWebVitals from './reportWebVitals';
import { LocalizationProvider } from '@mui/x-date-pickers';

import dayjs from 'dayjs';
dayjs.extend(customParseFormat); // if needed
dayjs.extend(localizedFormat); // if needed
dayjs.locale(i18n.language);

// ==============================|| MAIN - REACT DOM RENDER  ||============================== //

const container = document.getElementById('root');
const root = createRoot(container); // createRoot(container!) if you use TypeScript
root.render(
    <StrictMode>
        <ReduxProvider store={store}>
            <BrowserRouter basename="/">    
                <LocalizationProvider dateAdapter={AdapterDayjs} adapterLocale={i18n.language}>
                    <App />
                </LocalizationProvider>
            </BrowserRouter>
        </ReduxProvider>
    </StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
