// ==============================|| OVERRIDES - TOOLTIP ||============================== //

export default function Tooltip(theme) {
    const isDark = theme.palette.mode === 'dark';

    return {
        MuiTooltip: {
            defaultProps: {
                arrow: false,
            },
            styleOverrides: {
                tooltip: {
                    backgroundColor: isDark ? '#1e222b' : '#3a3a4a',
                    color: '#ffffff',
                    fontSize: '0.8125rem',
                    fontWeight: 500,
                    padding: '8px 12px',
                    borderRadius: '8px',
                    border: isDark ? '1px solid rgba(255, 255, 255, 0.12)' : '1px solid rgba(0, 0, 0, 0.12)',
                    boxShadow: '0 4px 12px rgba(0, 0, 0, 0.3)',
                    maxWidth: 300,
                    backdropFilter: 'none',
                },
                arrow: {
                    color: isDark ? '#1e222b' : '#3a3a4a',
                },
                popper: {
                    '& .MuiTooltip-tooltip': {
                        backgroundColor: isDark ? '#1e222b' : '#3a3a4a',
                    },
                },
            },
        },
    };
}
