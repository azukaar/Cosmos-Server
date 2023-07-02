// project import
import Routes from './routes';
import * as React from 'react';
import ThemeCustomization from './themes';
import ScrollTop from './components/ScrollTop';
import Snackbar from '@mui/material/Snackbar';
import {Alert, Box} from '@mui/material';
import logo from './assets/images/icons/cosmos.png';

import * as API from './api';

import { setSnackit, snackit } from './api/wrap';
import { DisconnectOutlined } from '@ant-design/icons';

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
    const [timeoutError, setTimeoutError] = React.useState(false);

    SetPrimaryColor = (color) => {
        setPrimaryColor(color);
        GlobalPrimaryColor = color;
    }

    SetSecondaryColor = (color) => {
        setSecondaryColor(color);
        GlobalSecondaryColor = color;
    }

    React.useEffect(() => {
        const timeout = setTimeout(
            () => {
                setTimeoutError(true);
            }, 10000
        )
        API.getStatus(true).then((r) => {
            clearTimeout(timeout);
            if(r == "NOT_AVAILABLE") {
                setTimeoutError(true);
            }
            else if(r) {
                setStatusLoaded(true);
            }
            setPrimaryColor(API.PRIMARY_COLOR);
            setSecondaryColor(API.SECONDARY_COLOR);

        }).catch(() => {
            clearTimeout(timeout);
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
                {!timeoutError && <LoadingAnimation />}
                {timeoutError && <DisconnectOutlined style={{
                    fontSize: '200px',
                    color: 'red',
                }}/>}
            </Box>
        </div>

}

export default App;
