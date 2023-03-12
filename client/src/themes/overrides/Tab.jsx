// ==============================|| OVERRIDES - TAB ||============================== //

export default function Tab(theme) {
    return {
        MuiTab: {
            styleOverrides: {
                root: {
                    minHeight: 46,
                    color: theme.palette.text.primary
                }
            }
        }
    };
}
