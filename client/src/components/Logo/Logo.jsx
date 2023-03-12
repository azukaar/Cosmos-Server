// material-ui
import { useTheme } from '@mui/material/styles';
import { fontWeight } from '@mui/system';

import logo from '../../../../Logo.png';

// ==============================|| LOGO SVG ||============================== //

const Logo = () => {
    const theme = useTheme();

    return (
        /**
         * if you want to use image instead of svg uncomment following, and comment out <svg> element.
         *
         * <img src={logo} alt="Mantis" width="100" />
         *
         */
        <>
            <img src={logo} alt="Cosmos" width="50" />
            <span style={{fontWeight: 'bold', fontSize: '170%', paddingLeft:'10px'}}> Cosmos</span>
        </>
    );
};

export default Logo;
