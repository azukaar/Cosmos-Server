// ==============================|| PRESET THEME - THEME SELECTOR ||============================== //

import { purple, pink, deepPurple, blueGrey, lightBlue, lightGreen, orange, teal, indigo } from '@mui/material/colors';

const Theme = (colors, darkMode) => {
    const { blue, red, gold, cyan, green, grey } = colors;
    const greyColors = {
        0: grey[0],
        50: grey[1],
        100: grey[2],
        200: grey[3],
        300: grey[4],
        400: grey[5],
        500: grey[6],
        600: grey[7],
        700: grey[8],
        800: grey[9],
        900: grey[10],
        A50: grey[15],
        A100: grey[11],
        A200: grey[12],
        A400: grey[13],
        A700: grey[14],
        A800: grey[16]
    };
    const contrastText = '#fff';

    return {
        primary: {
            main: purple[400],
        },
        secondary: {
            main: indigo[700],
        },
        error: {
            lighter: red[0],
            light: red[2],
            main: red[4],
            dark: red[7],
            darker: red[9],
            contrastText
        },
        warning: {
            lighter: gold[0],
            light: gold[3],
            main: gold[5],
            dark: gold[7],
            darker: gold[9],
            contrastText: greyColors[100]
        },
        info: {
            lighter: cyan[0],
            light: cyan[3],
            main: cyan[5],
            dark: cyan[7],
            darker: cyan[9],
            contrastText
        },
        success: {
            lighter: green[0],
            light: green[3],
            main: green[5],
            dark: green[7],
            darker: green[9],
            contrastText
        },
        debug: {
          lighter: grey[0],
          light: grey[3],
          main: grey[5],
          dark: grey[7],
          darker: grey[9],
          contrastText
        },
        important: {
          lighter: pink['100'],
          light: pink['200'],
          main: pink['400'],
          dark: pink['700'],
          darker: pink['800'],
          contrastText
        },
        grey: greyColors
    };
};

export default Theme;
