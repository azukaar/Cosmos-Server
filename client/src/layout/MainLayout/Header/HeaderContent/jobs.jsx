import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

// material-ui
import { useTheme } from '@mui/material/styles';
import {
    Avatar,
    Badge,
    Box,
    ClickAwayListener,
    Divider,
    IconButton,
    List,
    ListItemButton,
    ListItemAvatar,
    ListItemText,
    ListItemSecondaryAction,
    Paper,
    Popper,
    Typography,
    useMediaQuery,
    Button
} from '@mui/material';
import { register, format } from 'timeago.js';
import de from "timeago.js/lib/lang/de";

// project import
import MainCard from '../../../../components/MainCard';
import Transitions from '../../../../components/@extended/Transitions';
// assets
import { BellOutlined, ClockCircleOutlined, CloseOutlined, CloseSquareOutlined, ExclamationCircleOutlined, GiftOutlined, InfoCircleOutlined, LoadingOutlined, MessageOutlined, PlaySquareOutlined, SettingOutlined, WarningOutlined } from '@ant-design/icons';

import * as API from '../../../../api';
import { redirectToLocal } from '../../../../utils/indexs';
import { useClientInfos } from '../../../../utils/hooks';

// sx styles
const avatarSX = {
    width: 36,
    height: 36,
    fontSize: '1rem'
};

const actionSX = {
    mt: '6px',
    ml: 1,
    top: 'auto',
    right: 'auto',
    alignSelf: 'flex-start',

    transform: 'none'
};

// ==============================|| HEADER CONTENT - NOTIFICATION ||============================== //
const getStatus = (job) => {
    if (job.Running) return 'running';
    if (job.LastRunSuccess) return 'success';
    if (job.LastRun === "0001-01-01T00:00:00Z") return 'never';
    return 'error';
}

const Jobs = () => {
    register('de', de);
    const { t, i18n } = useTranslation();
    const {role} = useClientInfos();
    const isAdmin = role === "2";
    const theme = useTheme();
    const matchesXs = useMediaQuery(theme.breakpoints.down('md'));
    const [jobs, setJobs] = useState([]);
    const [from, setFrom] = useState('');
    
    const statusColor = (job) => {
        return {
            running: theme.palette.info.main,
            success: theme.palette.success.main,
            never: theme.palette.grey[500],
            error: theme.palette.error.main
        }[getStatus(job)]
    }

    const refreshJobs = () => {
        API.cron.list().then((res) => {
            setJobs(() => res.data);
        });
    };

    useEffect(() => {
        if (!isAdmin) return;

        refreshJobs();

        const interval = setInterval(() => {
            refreshJobs();
        }, 15000);

        return () => clearInterval(interval);
    }, []);

    const anchorRef = useRef(null);
    const [open, setOpen] = useState(false);
    
    const handleToggle = () => {
        refreshJobs();
        setOpen((prevOpen) => !prevOpen);
    };

    const handleClose = (event) => {
        if (anchorRef.current && anchorRef.current.contains(event.target)) {
            return;
        }
        setOpen(false);
    };

    const iconBackColor = theme.palette.mode === 'dark' ? 'grey.700' : 'grey.100';
    const iconBackColorOpen = theme.palette.mode === 'dark' ? 'grey.800' : 'grey.200';

    const nbRunning = 0

    let flatList = [];
    jobs && Object.values(jobs).forEach((list) => list && Object.values(list).forEach(job => flatList.push(job)));
    flatList.sort((a, b) => {
        if (a.Running && !b.Running) return -1;
        if (!a.Running && b.Running) return 1;
        if (a.Running && b.Running) return a.LastStarted - b.LastStarted;
        return a.LastRun - b.LastRun;
    });

    if (!isAdmin) return null;

    return (
        <Box sx={{ flexShrink: 0, ml: 0.75 }}>
            <IconButton
                disableRipple
                color="secondary"
                sx={{ color: 'text.primary', bgcolor: open ? iconBackColorOpen : iconBackColor }}
                aria-label="open profile"
                ref={anchorRef}
                aria-controls={open ? 'profile-grow' : undefined}
                aria-haspopup="true"
                onClick={handleToggle}
            >
                <Badge badgeContent={nbRunning} color="error">
                    <ClockCircleOutlined />
                </Badge>
            </IconButton>
            <Popper
                placement={matchesXs ? 'bottom' : 'bottom-end'}
                open={open}
                anchorEl={anchorRef.current}
                role={undefined}
                transition
                disablePortal
                popperOptions={{
                    modifiers: [
                        {
                            name: 'offset',
                            options: {
                                offset: [matchesXs ? -5 : 0, 9]
                            }
                        }
                    ]
                }}
            >
                {({ TransitionProps }) => (
                    <Transitions type="fade" in={open} {...TransitionProps}>
                        <Paper
                            sx={{
                                boxShadow: theme.customShadows.z1,
                                width: '100%',
                                minWidth: 350,
                                maxWidth: 400,
                                [theme.breakpoints.down('md')]: {
                                    maxWidth: 285
                                }
                            }}
                        >
                            <ClickAwayListener onClickAway={handleClose}>
                                <MainCard
                                    title="Jobs"
                                    elevation={0}
                                    border={false}
                                    content={false}
                                    secondary={
                                        <IconButton size="small" onClick={handleToggle}>
                                            <CloseOutlined />
                                        </IconButton>
                                    }
                                >
                                    <List
                                        component="nav"
                                        sx={{
                                            maxHeight: 350,
                                            overflow: 'auto',
                                            p: 0,
                                            '& .MuiListItemButton-root': {
                                                py: 0.5,
                                                '& .MuiAvatar-root': avatarSX,
                                                '& .MuiListItemSecondaryAction-root': { ...actionSX, position: 'relative' }
                                            }
                                        }}
                                    >
                                        {flatList && flatList.map(job => (<>
                                            <ListItemButton onClick={() => {
                                                job.Link && redirectToLocal(job.Link);
                                            }}
                                            style={{
                                                borderLeft: job.Read ? 'none' : `4px solid ${statusColor(job)}`,
                                                paddingLeft: job.Read ? '14px' : '10px',
                                            }}>
                                            
                                            <ListItemText
                                                primary={<>
                                                    <Typography variant={job.Read ? 'body' : 'h6'} noWrap>
                                                        {job.Name}
                                                    </Typography>
                                                    <div style={{ 
                                                        overflow: 'hidden',
                                                        maxHeight: '48px',
                                                        borderLeft: '1px solid grey',
                                                        paddingLeft: '8px',
                                                        margin: '2px'
                                                    }}>
                                                        {job.Message}
                                                    </div></>
                                                }
                                            />
                                                <ListItemSecondaryAction>
                                                    
                                                {(job.Running) ? <IconButton disabled={!job.Cancellable} variant="outlined" color="error" onClick={() => {
                                                    API.cron.stop(job.Scheduler, job.Name).then(() => {
                                                        refreshJobs();
                                                    });
                                                }}>
                                                    <CloseSquareOutlined />
                                                </IconButton> : ''}
                                                {(!job.Running) ? <IconButton variant="outlined" color="primary" onClick={() => {
                                                    API.cron.run(job.Scheduler, job.Name).then(() => {
                                                        refreshJobs();
                                                    });
                                                }}>
                                                    <PlaySquareOutlined />
                                                </IconButton> : ''}
                                                </ListItemSecondaryAction>
                                            </ListItemButton>
                                            <ListItemButton
                                            style={{
                                                borderLeft: job.Read ? 'none' : `4px solid  ${statusColor(job)}`,
                                                paddingLeft: job.Read ? '14px' : '10px',
                                                display: 'flex',
                                                justifyContent: 'flex-end',
                                                alignItems: 'center',
                                            }}>
                                            <Typography variant="caption" noWrap >
                                                {job.LastStarted == '0001-01-01T00:00:00Z' ? t('mgmt.scheduler.list.status.neverRan') : (
                                                    job.Running ? <span><LoadingOutlined />{` `+t('mgmt.cron.list.state.running')+` ${format(job.LastStarted, i18n.resolvedLanguage)}`}</span> : t('mgmt.cron.list.state.lastRan')+` ${format(job.LastRun, i18n.resolvedLanguage)}`
                                                )}
                                            </Typography>
                                            </ListItemButton>
                                        <Divider /></>))}
                                    </List>
                                </MainCard>
                            </ClickAwayListener>
                        </Paper>
                    </Transitions>
                )}
            </Popper>
        </Box>
    );
};

export default Jobs;
