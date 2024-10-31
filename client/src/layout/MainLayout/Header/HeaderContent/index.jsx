import React from 'react';

// material-ui
import { Box, Chip, IconButton, Link, Stack, useMediaQuery } from '@mui/material';
import { GithubOutlined } from '@ant-design/icons';

// project import
import Search from './Search';
import Profile from './Profile';
import Notification from './Notification';
import MobileSection from './MobileSection';
import Jobs from './jobs';
import { useTranslation } from 'react-i18next';
import RestartMenu from './restartMenu';
import TerminalHeader from './terminal';

// ==============================|| HEADER - CONTENT ||============================== //

const HeaderContent = () => {
    const { t } = useTranslation();
    const matchesXs = useMediaQuery((theme) => theme.breakpoints.down('md'));

    return (
        <>
            {!matchesXs && <Search />}
            {matchesXs && <Box sx={{ width: '100%', ml: 1 }} />}

            <Stack direction="row" spacing={2}>
                <TerminalHeader />
                <Jobs />
                <Notification />

                <Link href="/cosmos-ui/logout" underline="none">
                    <Chip label={t('global.logout')} style={{height: '36px'}} />
                </Link>
                <RestartMenu />
            </Stack>
            {/* {!matchesXs && <Profile />}
            {matchesXs && <MobileSection />} */}
        </>
    );
};

export default HeaderContent;
