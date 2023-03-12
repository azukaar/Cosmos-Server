// ==============================|| OVERRIDES - CARD CONTENT ||============================== //

export default function CardContent() {
    return {
        MuiCardContent: {
            styleOverrides: {
                root: {
                    padding: 20,
                    '&:last-child': {
                        paddingBottom: 20
                    }
                }
            }
        }
    };
}
