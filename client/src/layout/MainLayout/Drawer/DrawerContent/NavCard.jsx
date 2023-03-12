// material-ui
import { Button, CardMedia, Link, Stack, Typography } from '@mui/material';

// project import
import MainCard from '../../../../components/MainCard';

// assets
import avatar from '../../../../assets/images/users/avatar-group.png';
import AnimateButton from '../../../../components/@extended/AnimateButton';

// ==============================|| DRAWER CONTENT - NAVIGATION CARD ||============================== //

const NavCard = () => (
    <MainCard sx={{ bgcolor: 'grey.50', m: 3 }}>
        <Stack alignItems="center" spacing={2.5}>
            <Stack alignItems="center">
                <Typography variant="h5">Cosmos Pro</Typography>
                <Typography variant="h6" color="secondary">
                    Checkout pro features
                </Typography>
            </Stack>
            <AnimateButton>
                <Button component={Link} target="_blank" href="https://mantisdashboard.io" variant="contained" color="success" size="small">
                    Pro
                </Button>
            </AnimateButton>
        </Stack>
    </MainCard>
);

export default NavCard;
