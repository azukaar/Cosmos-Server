// project import
import Routes from './routes';
import * as React from 'react';
import ThemeCustomization from './themes';
import ScrollTop from './components/ScrollTop';
import Snackbar from '@mui/material/Snackbar';
import {Alert, Box} from '@mui/material';
import logo from './assets/images/icons/cosmos.png';

import * as API from './api';

import { setSnackit } from './api/wrap';

// ==============================|| APP - THEME, ROUTER, LOCAL  ||============================== //

const LoadingAnimation = () => (
    <div className="loader">
      <div className="dot"></div>
      <div className="dot"></div>
      <div className="dot"></div>
    </div>
  );

export let SetPrimaryColor = () => {};
export let SetSecondaryColor = () => {};
export let GlobalPrimaryColor = '';
export let GlobalSecondaryColor = '';
  
const App = () => {
    const [open, setOpen] = React.useState(false);
    const [message, setMessage] = React.useState('');
    const [severity, setSeverity] = React.useState('error');
    const [statusLoaded, setStatusLoaded] = React.useState(false);
    const [PrimaryColor, setPrimaryColor] = React.useState(API.PRIMARY_COLOR);
    const [SecondaryColor, setSecondaryColor] = React.useState(API.SECONDARY_COLOR);

    SetPrimaryColor = (color) => {
        setPrimaryColor(color);
        GlobalPrimaryColor = color;
    }

    SetSecondaryColor = (color) => {
        setSecondaryColor(color);
        GlobalSecondaryColor = color;
    }

    React.useEffect(() => {
        API.getStatus(true).then((r) => {
            if(r) {
                setStatusLoaded(true);
            }    
            setPrimaryColor(API.PRIMARY_COLOR);
            setSecondaryColor(API.SECONDARY_COLOR);
        }).catch(() => {
            setStatusLoaded(true);
            setPrimaryColor(API.PRIMARY_COLOR);
            setSecondaryColor(API.SECONDARY_COLOR);
        });
    }, []);

    setSnackit((message, severity='error') => {
        setMessage(message);
        setOpen(true);
        setSeverity(severity);
    })

    return statusLoaded ?
        <ThemeCustomization PrimaryColor={PrimaryColor} SecondaryColor={SecondaryColor}>
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

        : <div>
            <Box sx={{ position: 'fixed', top: 0, bottom: 0, left: 0, right: 0, display: 'flex', justifyContent: 'center', alignItems: 'center'}}>
                {/* <img src={logo} style={{ display:'inline', height: '200px'}} className='pulsing' /> */}
                <LoadingAnimation />
            </Box>
        </div>

}

export default App;
