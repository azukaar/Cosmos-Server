import PropTypes from 'prop-types';
import { forwardRef } from 'react';

// material-ui
import { Fade, Box, Grow } from '@mui/material';

// ==============================|| TRANSITIONS ||============================== //

const Transitions = forwardRef(({ children, position, type, ...others }, ref) => {
    let positionSX = {
        transformOrigin: '0 0 0'
    };

    switch (position) {
        case 'top-right':
        case 'top':
        case 'bottom-left':
        case 'bottom-right':
        case 'bottom':
        case 'top-left':
        default:
            positionSX = {
                transformOrigin: '0 0 0'
            };
            break;
    }

    return (
        <Box ref={ref}>
            {type === 'grow' && (
                <Grow {...others}>
                    <Box sx={positionSX}>{children}</Box>
                </Grow>
            )}
            {type === 'fade' && (
                <Fade
                    {...others}
                    timeout={{
                        appear: 0,
                        enter: 300,
                        exit: 150
                    }}
                >
                    <Box sx={positionSX}>{children}</Box>
                </Fade>
            )}
        </Box>
    );
});

Transitions.propTypes = {
    children: PropTypes.node,
    type: PropTypes.oneOf(['grow', 'fade', 'collapse', 'slide', 'zoom']),
    position: PropTypes.oneOf(['top-left', 'top-right', 'top', 'bottom-left', 'bottom-right', 'bottom'])
};

Transitions.defaultProps = {
    type: 'grow',
    position: 'top-left'
};

export default Transitions;
