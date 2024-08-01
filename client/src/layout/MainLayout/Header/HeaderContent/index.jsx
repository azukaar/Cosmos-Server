// material-ui
import { Box, Chip, IconButton, Link, Stack, useMediaQuery } from '@mui/material';
import { GithubOutlined } from '@ant-design/icons';

// project import
import Search from './Search';
import Profile from './Profile';
import Notification from './Notification';
import MobileSection from './MobileSection';
import Jobs from './jobs';

// ==============================|| HEADER - CONTENT ||============================== //

const HeaderContent = () => {
    const matchesXs = useMediaQuery((theme) => theme.breakpoints.down('md'));

    return (
        <>
            {!matchesXs && <Search />}
            {matchesXs && <Box sx={{ width: '100%', ml: 1 }} />}

            <Stack direction="row" spacing={2}>
                <Jobs />
                <Notification />

                <Link href="/cosmos-ui/logout" underline="none">
                    <Chip label="Logout" />
                </Link>
            </Stack>
            {/* {!matchesXs && <Profile />}
            {matchesXs && <MobileSection />} */}
        </>
    );
};

export default HeaderContent;
