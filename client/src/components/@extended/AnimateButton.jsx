import PropTypes from 'prop-types';

// third-party
import { motion } from 'framer-motion';

// ==============================|| ANIMATION BUTTON ||============================== //

export default function AnimateButton({ children, type }) {
    switch (type) {
        case 'rotate': // only available in paid version
        case 'slide': // only available in paid version
        case 'scale': // only available in paid version
        default:
            return (
                <motion.div whileHover={{ scale: 1 }} whileTap={{ scale: 0.9 }}>
                    {children}
                </motion.div>
            );
    }
}

AnimateButton.propTypes = {
    children: PropTypes.node,
    type: PropTypes.oneOf(['slide', 'scale', 'rotate'])
};

AnimateButton.defaultProps = {
    type: 'scale'
};
