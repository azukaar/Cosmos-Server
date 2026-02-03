// material-ui
import { alpha } from '@mui/material/styles';

// ==============================|| OVERRIDES - BUTTON ||============================== //

export default function Button(theme) {
    const primaryMain = theme.palette.primary.main;

    const disabledStyle = {
        '&.Mui-disabled': {
            backgroundColor: theme.palette.grey[400]
        }
    };

    return {
        MuiButton: {
            defaultProps: {
                disableElevation: true
            },
            styleOverrides: {
                root: {
                    fontWeight: 500,
                    borderRadius: 8,
                    textTransform: 'none',
                    transition: 'all 0.15s ease',
                },
                contained: {
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: 'none',
                        filter: 'brightness(1.1)',
                    },
                    ...disabledStyle
                },
                outlined: {
                    borderWidth: '1.5px',
                    '&:hover': {
                        borderWidth: '1.5px',
                        backgroundColor: alpha(primaryMain, 0.08),
                    },
                    ...disabledStyle
                },
                text: {
                    '&:hover': {
                        backgroundColor: alpha(primaryMain, 0.06),
                    },
                },
            }
        }
    };
}
