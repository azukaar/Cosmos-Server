// project import
import Routes from './routes';
import * as React from 'react';
import ThemeCustomization from './themes';
import ScrollTop from './components/ScrollTop';
import Snackbar from '@mui/material/Snackbar';
import {Alert} from '@mui/material';

import { setSnackit } from './api/wrap';

// ==============================|| APP - THEME, ROUTER, LOCAL  ||============================== //

const App = () => {
    const [open, setOpen] = React.useState(false);
    const [message, setMessage] = React.useState('');
    const [severity, setSeverity] = React.useState('error');
    setSnackit((message, severity='error') => {
        setMessage(message);
        setOpen(true);
        setSeverity(severity);
    })
    return (
        <ThemeCustomization>
            <Snackbar
                open={open}
                autoHideDuration={5000}
                onClose={() => {setOpen(false)}}
                anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
            >  
                <Alert className={(open && severity == "error") ? 'shake' : ''} severity={severity} sx={{ width: '100%' }}>
                    {message}
                </Alert>
            </Snackbar>
            <ScrollTop>
                <Routes />
            </ScrollTop>
        </ThemeCustomization>
    )
}

export default App;
