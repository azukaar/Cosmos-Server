// material-ui
import { alpha } from '@mui/material/styles';

// ==============================|| DEFAULT THEME - CUSTOM SHADOWS  ||============================== //

const CustomShadows = (theme) => ({
    button: `0 2px #0000000b`,
    text: `0 -1px 0 rgb(0 0 0 / 12%)`,
    z1: `0px 2px 8px ${alpha(theme.palette.grey[900], 0.15)}`,
    glass: `0 4px 24px ${alpha(theme.palette.grey[900], 0.12)}, 0 1px 2px ${alpha(theme.palette.grey[900], 0.08)}`,
    glassHover: `0 8px 32px ${alpha(theme.palette.grey[900], 0.22)}`,
    cardFloat: `0 12px 40px ${alpha(theme.palette.grey[900], 0.15)}`
});

export default CustomShadows;
