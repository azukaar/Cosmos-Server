// ==============================|| OVERRIDES - CHIP ||============================== //

export default function Chip(theme) {
    const primaryMain = theme.palette.primary.main;
    const secondaryMain = theme.palette.secondary.main;

    const colorGradient = (name) => {
        const c = theme.palette[name];
        return `linear-gradient(135deg, ${c.main}, ${c.dark})`;
    };

    return {
        MuiChip: {
            styleOverrides: {
                root: {
                    borderRadius: 8,
                    '&:active': {
                        boxShadow: 'none'
                    }
                },
                outlinedWarning: {
                    color: theme.palette.warning.dark,
                },
                sizeLarge: {
                    fontSize: '1rem',
                    height: 40
                },
                // --- Filled chips: gradient backgrounds ---
                filledPrimary: {
                    backgroundImage: `linear-gradient(135deg, ${primaryMain}, ${secondaryMain})`,
                },
                filledSecondary: {
                    backgroundImage: colorGradient('secondary'),
                },
                filledError: {
                    backgroundImage: colorGradient('error'),
                },
                filledSuccess: {
                    backgroundImage: colorGradient('success'),
                },
                filledWarning: {
                    backgroundImage: colorGradient('warning'),
                    color: '#000',
                },
                filledInfo: {
                    backgroundImage: colorGradient('info'),
                },
                // --- Light variant (custom) ---
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
