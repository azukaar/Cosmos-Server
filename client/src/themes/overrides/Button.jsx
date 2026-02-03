// ==============================|| OVERRIDES - BUTTON ||============================== //

export default function Button(theme) {
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
                    backdropFilter: 'blur(4px)',
                    transition: 'all 0.2s ease',
                },
                contained: {
                    ...disabledStyle
                },
                outlined: {
                    ...disabledStyle
                }
            }
        }
    };
}
