import PropTypes from 'prop-types';
import { useMemo } from 'react';

// material-ui
import { useTheme } from '@mui/material/styles';
import { Box, Chip, Drawer, useMediaQuery } from '@mui/material';

// project import
import DrawerHeader from './DrawerHeader';
import DrawerContent from './DrawerContent';
import MiniDrawerStyled from './MiniDrawerStyled';
import { drawerWidth } from '../../../config';
import NavCard from './DrawerContent/NavCard';

import {version} from '../../../../../package.json';
import { LanguagesSelect } from './languages';

// ==============================|| MAIN LAYOUT - DRAWER ||============================== //

const MainDrawer = ({ open, handleDrawerToggle, window }) => {
    const theme = useTheme();
    const matchDownMD = useMediaQuery(theme.breakpoints.down('lg'));

    // responsive drawer container
    const container = window !== undefined ? () => window().document.body : undefined;

    // header content
    const drawerContent = useMemo(() => <DrawerContent />, []);
    const drawerHeader = useMemo(() => <DrawerHeader open={open} />, [open]);
    const footer = <div style={{ margin: '10px', display: 'flex', justifyContent: 'center' }}>
        <LanguagesSelect />
        <Chip
            label={version.replace('unstable', '')}
            sx={{ marginLeft: '10px', height: 24, '& .MuiChip-label': { fontSize: '0.7rem', py: 0.25 } }}
            component="a"
            href="/"
            clickable
        />
    </div>;

    return (
        <Box component="nav" sx={{ flexShrink: { md: 0 }, zIndex: 1300 }} aria-label="mailbox folders">
            {!matchDownMD ? (
                <MiniDrawerStyled
                    variant="permanent" open={open}>
                    {drawerHeader}
                    {drawerContent}
                    {footer}
                </MiniDrawerStyled>
            ) : (
                <Drawer
                    container={container}
                    variant="temporary"
                    open={open}
                    onClose={handleDrawerToggle}
                    ModalProps={{ keepMounted: true }}
                    sx={{
                        display: { xs: 'block', lg: 'none' },
                        '& .MuiDrawer-paper': {
                            boxSizing: 'border-box',
                            width: drawerWidth,
                            borderRight: `1px solid ${theme.palette.divider}`,
                            backgroundImage: 'none',
                            boxShadow: 'inherit'
                        }
                    }}
                >
                    {open && drawerHeader}
                    {open && drawerContent}
                    {open && footer}
                </Drawer>
            )}
        </Box>
    );
};

MainDrawer.propTypes = {
    open: PropTypes.bool,
    handleDrawerToggle: PropTypes.func,
    window: PropTypes.object
};

export default MainDrawer;
