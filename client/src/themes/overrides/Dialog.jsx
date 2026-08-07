// ==============================|| OVERRIDES - DIALOG ||============================== //

export default function Dialog(theme) {
    const isDark = theme.palette.mode === 'dark';

    return {
        MuiBackdrop: {
            styleOverrides: {
                root: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                },
            },
        },
        MuiDialog: {
            styleOverrides: {
                paper: {
                    backgroundColor: isDark ? 'rgba(32, 35, 43, 1)' : '#ffffff',
                    backgroundImage: 'none',
                    borderRadius: 12,
                    border: isDark ? '1px solid rgba(255, 255, 255, 0.08)' : '1px solid rgba(0, 0, 0, 0.06)',
                    boxShadow: isDark
                        ? '0 8px 32px rgba(0, 0, 0, 0.4)'
                        : '0 8px 32px rgba(0, 0, 0, 0.12)',
                },
            },
        },
        MuiDialogTitle: {
            styleOverrides: {
                root: {
                    fontSize: '1.125rem',
                    fontWeight: 600,
                },
            },
        },
        MuiDialogContent: {
            styleOverrides: {
                root: {
                    padding: '16px 24px',
                },
            },
        },
        MuiDialogActions: {
            styleOverrides: {
                root: {
                    padding: '16px 24px',
                },
            },
        },
    };
}
