import PropTypes from 'prop-types';

// material-ui
import { useTheme } from '@mui/material/styles';
import { Stack, Chip } from '@mui/material';

// project import
import DrawerHeaderStyled from './DrawerHeaderStyled';
import Logo from '../../../../components/Logo';
import {version} from '../../../../../../package.json';

// ==============================|| DRAWER HEADER ||============================== //

const DrawerHeader = ({ open }) => {
    const theme = useTheme();

    return (
        <DrawerHeaderStyled theme={theme} open={open}>
            <Stack direction="row" spacing={1} alignItems="center">
                <Logo />
                <Chip
                    label={version.replace('unstable', '')}
                    size="small"
                    sx={{ height: 16, '& .MuiChip-label': { fontSize: '0.55rem', py: 0.25 } }}
                    component="a"
                    href="/"
                    clickable
                />
            </Stack>
        </DrawerHeaderStyled>
    );
};

DrawerHeader.propTypes = {
    open: PropTypes.bool
};

export default DrawerHeader;
