// material-ui
import { Box, Typography } from '@mui/material';

// project import
import NavGroup from './NavGroup';
import menuItem from '../../../../../menu-items';
import { useIsPro } from '../../../../../utils/pro';

// ==============================|| DRAWER CONTENT - NAVIGATION ||============================== //

const Navigation = () => {
    // Renames entries Pro overrides via `titlePro`; the group shape is set in menu-items/index.jsx.
    const isPro = useIsPro() === true;

    const navGroups = menuItem.items
        .map((item) => ({
            ...item,
            children: (item.children || []).map((child) =>
                (isPro && child.titlePro ? { ...child, title: child.titlePro } : child)),
        }))
        .map((item) => {
            switch (item.type) {
                case 'group':
                    return <NavGroup key={item.id} item={item} />;
                default:
                    return (
                        <Typography key={item.id} variant="h6" color="error" align="center">
                            Fix - Navigation Group
                        </Typography>
                    );
            }
        });

    return <Box sx={{ pt: 2 }}>{navGroups}</Box>;
};

export default Navigation;
