// ==============================|| OVERRIDES - CHIP ||============================== //

export default function Chip(theme) {
    return {
        MuiChip: {
            styleOverrides: {
                root: {
                    borderRadius: 4,
                    '&:active': {
                        boxShadow: 'none'
                    }
                },
                sizeLarge: {
                    fontSize: '1rem',
                    height: 40
                },
                light: {
                    color: theme.palette.primary.main,
                    backgroundColor: theme.palette.primary.lighter,
                    borderColor: theme.palette.primary.light,
                    '&.MuiChip-lightError': {
                        color: theme.palette.error.main,
                        backgroundColor: theme.palette.error.lighter,
                        borderColor: theme.palette.error.light
                    },
                    '&.MuiChip-lightSuccess': {
                        color: theme.palette.success.main,
                        backgroundColor: theme.palette.success.lighter,
                        borderColor: theme.palette.success.light
                    },
                    '&.MuiChip-lightWarning': {
                        color: theme.palette.warning.main,
                        backgroundColor: theme.palette.warning.lighter,
                        borderColor: theme.palette.warning.light
                    }
                }
            }
        }
    };
}
