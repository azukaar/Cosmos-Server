// material-ui
import { alpha } from '@mui/material/styles';

// ==============================|| OVERRIDES - OUTLINED INPUT ||============================== //

export default function OutlinedInput(theme) {
    return {
        MuiOutlinedInput: {
            styleOverrides: {
                input: {
                    padding: '10.5px 14px 10.5px 12px'
                },
                notchedOutline: {
                    borderColor: theme.palette.grey[300]
                },
                root: {
                    '&:hover .MuiOutlinedInput-notchedOutline': {
                        borderColor: theme.palette.primary.light
                    },
                    '&.Mui-focused': {
                        boxShadow: `0 0 0 2px ${alpha(theme.palette.primary.main, 0.2)}`,
                        '& .MuiOutlinedInput-notchedOutline': {
                            border: `1px solid ${theme.palette.primary.light}`
                        }
                    },
                    '&.Mui-error': {
                        '&:hover .MuiOutlinedInput-notchedOutline': {
                            borderColor: theme.palette.error.light
                        },
                        '&.Mui-focused': {
                            boxShadow: `0 0 0 2px ${alpha(theme.palette.error.main, 0.2)}`,
                            '& .MuiOutlinedInput-notchedOutline': {
                                border: `1px solid ${theme.palette.error.light}`
                            }
                        }
                    }
                },
                inputSizeSmall: {
                    padding: '7.5px 8px 7.5px 12px'
                },
                inputMultiline: {
                    padding: 0
                }
            }
        }
    };
}
