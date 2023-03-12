// material-ui
import { useTheme } from '@mui/material/styles';
import { Box } from '@mui/material';

import logo from '../../../../../Logo.png';

// ==============================|| AUTH BLUR BACK SVG ||============================== //

const AuthBackground = () => {
    const theme = useTheme();
    return (
        <Box sx={{ position: 'fixed', float: 'left', height: 'calc(100vh - 50px)', overflow: 'hidden', filter: 'blur(25px)', zIndex: -1, top: 100, left: -500 }}>
            <img src={logo} style={{ display:'inline'}} alt="Cosmos" width="1100" />
        </Box>
    );
};

export default AuthBackground;