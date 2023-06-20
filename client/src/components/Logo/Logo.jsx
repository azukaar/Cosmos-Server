// material-ui
import { useTheme } from '@mui/material/styles';
import { fontWeight } from '@mui/system';

import logo from '../../assets/images/icons/cosmos_simple_black.png';
import logoDark from '../../assets/images/icons/cosmos_simple_white.png';

// ==============================|| LOGO SVG ||============================== //

const Logo = () => {
    const theme = useTheme();
    const isLight = theme.palette.mode === 'light';

    return (
        /**
         * if you want to use image instead of svg uncomment following, and comment out <svg> element.
         *
         * <img src={logo} alt="Mantis" width="100" />
         *
         */
        <>
            <img src={isLight ? logo : logoDark} alt="Cosmos" width="40" />
            <span style={{fontWeight: 'bold', fontSize: '170%', paddingLeft:'10px'}}> Cosmos</span>
        </>
    );
};

export default Logo;
