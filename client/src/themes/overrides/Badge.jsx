// ==============================|| OVERRIDES - BADGE ||============================== //

export default function Badge(theme) {
    const primaryMain = theme.palette.primary.main;
    const secondaryMain = theme.palette.secondary.main;

    const colorGradient = (name) => {
        const c = theme.palette[name];
        return `linear-gradient(135deg, ${c.main}, ${c.dark})`;
    };

    return {
        MuiBadge: {
            styleOverrides: {
                standard: {
                    minWidth: theme.spacing(2),
                    height: theme.spacing(2),
                    padding: theme.spacing(0.5)
                },
                colorPrimary: {
                    backgroundImage: `linear-gradient(135deg, ${primaryMain}, ${secondaryMain})`,
                },
                colorSecondary: {
                    backgroundImage: colorGradient('secondary'),
                },
                colorError: {
                    backgroundImage: colorGradient('error'),
                },
                colorSuccess: {
                    backgroundImage: colorGradient('success'),
                },
                colorWarning: {
                    backgroundImage: colorGradient('warning'),
                },
                colorInfo: {
                    backgroundImage: colorGradient('info'),
                },
            }
        }
    };
}
