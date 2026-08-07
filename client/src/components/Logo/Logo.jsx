// material-ui
import { useTheme } from '@mui/material/styles';
import { fontWeight } from '@mui/system';

import logo from '../../assets/images/icons/cosmos_simple_black.png';
import logoDark from '../../assets/images/icons/cosmos_simple_white.png';

import logoSvg from '../../assets/images/icons/logo.svg';

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
            <img
                src={logoSvg}
                alt="Cosmos"
                width="40"
                style={{
                filter: !isLight
                    ? "brightness(0) invert(1)" // white
                    : "brightness(0)" // black
                }}
            />
            <span style={{fontWeight: 500, fontSize: '160%', paddingLeft:'10px'}}>
                Cosmos{' '}
                <span style={{
                    background: `linear-gradient(to right, ${theme.palette.primary.main}, ${theme.palette.secondary.main})`,
                    WebkitBackgroundClip: 'text',
                    WebkitTextFillColor: 'transparent',
                    backgroundClip: 'text',
                }}>Cloud</span>
            </span>
        </>
    );
};

export default Logo;
