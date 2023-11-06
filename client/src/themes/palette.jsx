// material-ui
import { createTheme } from '@mui/material/styles';

// third-party
import { presetPalettes } from '@ant-design/colors';

// project import
import ThemeOption from './theme';

import * as API from '../api';

// ==============================|| DEFAULT THEME - PALETTE  ||============================== //

const Palette = (mode, PrimaryColor, SecondaryColor) => {
    const colors = presetPalettes;

    const greyPrimary = [
        '#ffffff',
        '#fafafa',
        '#f5f5f5',
        '#f0f0f0',
        '#d9d9d9',
        '#bfbfbf',
        '#8c8c8c',
        '#595959',
        '#262626',
        '#141414',
        '#000000'
    ];
    const greyAscent = ['#fafafa', '#bfbfbf', '#434343', '#1f1f1f'];
    const greyConstant = ['#fafafb', '#e6ebf1'];

    colors.grey = [...greyPrimary, ...greyAscent, ...greyConstant];

    const paletteColor = ThemeOption(colors, mode === 'dark');

    if(PrimaryColor) {
        paletteColor.primary.main = PrimaryColor;
    }
    if(SecondaryColor) {
        paletteColor.secondary.main = SecondaryColor;
    }

    return createTheme(mode === 'dark' ? {
        palette: {
            mode,
            common: {
                black: '#fff',
                white: '#000'
            },
            ...paletteColor,
            text: {
                primary: paletteColor.grey[0],
                secondary: paletteColor.grey[200],
                disabled: paletteColor.grey[300]
            },
            action: {
                disabled: paletteColor.grey[300]
            },
            divider: paletteColor.grey[600],
            background: {
                paper: paletteColor.grey[700],
                default: paletteColor.grey[800]
            }
        },
    } : {
        palette: {
            mode,
            common: {
                black: '#000',
                white: '#fff'
            },
            ...paletteColor,
            text: {
                primary: paletteColor.grey[700],
                secondary: paletteColor.grey[600],
                disabled: paletteColor.grey[500]
            },
            action: {
                disabled: paletteColor.grey[300]
            },
            divider: paletteColor.grey[200],
            background: {
                paper: paletteColor.grey[0],
                default: paletteColor.grey.A50
            }
        },
    });
};

export default Palette;
