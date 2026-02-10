// material-ui
import { alpha } from '@mui/material/styles';

// ==============================|| OVERRIDES - BUTTON ||============================== //

export default function Button(theme) {
    const primaryMain = theme.palette.primary.main;
    const secondaryMain = theme.palette.secondary.main;
    const paperBg = theme.palette.background.paper;

    // Two-tone gradient: primary → secondary
    const primaryGradient = `linear-gradient(135deg, ${primaryMain}, ${secondaryMain})`;

    // Per-color gradient: main → dark
    const colorGradient = (name) => {
        const c = theme.palette[name];
        return `linear-gradient(135deg, ${c.main}, ${c.dark})`;
    };

    const disabledStyle = {
        '&.Mui-disabled': {
            backgroundColor: theme.palette.grey[400],
            backgroundImage: 'none',
            borderColor: theme.palette.grey[400],
        }
    };

    // Build per-color contained + outlined overrides
    const colors = {
        Primary:   { gradient: primaryGradient, main: primaryMain, shift: secondaryMain },
        Secondary: { gradient: colorGradient('secondary'), main: secondaryMain },
        Error:     { gradient: colorGradient('error'), main: theme.palette.error.main },
        Success:   { gradient: colorGradient('success'), main: theme.palette.success.main },
        Warning:   { gradient: colorGradient('warning'), main: theme.palette.warning.main },
        Info:      { gradient: colorGradient('info'), main: theme.palette.info.main },
    };

    const containedOverrides = {};
    const outlinedOverrides = {};

    Object.entries(colors).forEach(([name, { gradient }]) => {
        containedOverrides[`contained${name}`] = {
            backgroundImage: gradient,
        };

        outlinedOverrides[`outlined${name}`] = {
            borderColor: 'transparent',
            backgroundImage: `linear-gradient(${paperBg}, ${paperBg}), ${gradient}`,
            backgroundOrigin: 'border-box',
            backgroundClip: 'padding-box, border-box',
            '&:hover': {
                borderColor: 'transparent',
                backgroundColor: 'transparent',
                backgroundImage: gradient,
                backgroundOrigin: 'border-box',
                backgroundClip: 'border-box',
                color: '#fff',
            },
        };
    });

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

                // --- Contained: gradient backgrounds ---
                contained: {
                    boxShadow: 'none',
                    '&:hover': {
                        boxShadow: 'none',
                        filter: 'brightness(1.15)',
                    },
                    ...disabledStyle
                },
                ...containedOverrides,

                // --- Outlined: gradient borders ---
                outlined: {
                    borderWidth: '1.5px',
                    '&:hover': {
                        borderWidth: '1.5px',
                    },
                    ...disabledStyle
                },
                ...outlinedOverrides,

                // --- Text: gradient tint on hover ---
                text: {
                    '&:hover': {
                        backgroundColor: alpha(primaryMain, 0.06),
                    },
                },
                textPrimary: {
                    '&:hover': {
                        backgroundImage: `linear-gradient(135deg, ${alpha(primaryMain, 0.06)}, ${alpha(secondaryMain, 0.06)})`,
                        backgroundColor: 'transparent',
                    },
                },
            }
        }
    };
}
