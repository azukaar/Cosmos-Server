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
    useMediaQuery
} from '@mui/material';
import { register, format } from 'timeago.js';
import de from "timeago.js/lib/lang/de";

// project import
import MainCard from '../../../../components/MainCard';
import Transitions from '../../../../components/@extended/Transitions';
// assets
import { BellOutlined, CloseOutlined, ExclamationCircleOutlined, GiftOutlined, InfoCircleOutlined, MessageOutlined, SettingOutlined, WarningOutlined } from '@ant-design/icons';

import * as API from '../../../../api';
import { redirectToLocal } from '../../../../utils/indexs';

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

const Notification = () => {
    register('de', de);
    const { t, i18n } = useTranslation();
    const theme = useTheme();
    const matchesXs = useMediaQuery(theme.breakpoints.down('md'));
    const [notifications, setNotifications] = useState([]);
    const [from, setFrom] = useState('');

    const refreshNotifications = () => {
        API.users.getNotifs(from).then((res) => {
            setNotifications(() => res.data);
        });
    };

    const setAsRead = () => {
        let unread = [];
        
       notifications.forEach((notif) => {
            if (!notif.Read) {
                unread.push(notif.ID);
            }
        })

        if (unread.length > 0) {
            API.users.readNotifs(unread);
        }
    }

    const setLocalAsRead = (id) => {
        let newN = notifications.map((notif) => {
            notif.Read = true;
            return notif;
        })
        setNotifications(newN);
    }

    useEffect(() => {
        refreshNotifications();

        const interval = setInterval(() => {
            refreshNotifications();
        }, 10000);

        return () => clearInterval(interval);
    }, []);

    const anchorRef = useRef(null);
    const [open, setOpen] = useState(false);
    
    const handleToggle = () => {
        if (!open) {
            setAsRead();
        }

        setOpen((prevOpen) => !prevOpen);
    };

    const handleClose = (event) => {
        if (anchorRef.current && anchorRef.current.contains(event.target)) {
            return;
        }
        setLocalAsRead();
        setOpen(false);
    };

    const getNotifIcon = (notification) => {
        switch (notification.Level) {
            case 'warn':
                return <Avatar
                    sx={{
                        color: 'warning.main',
                        bgcolor: 'warning.lighter'
                    }}
                >
                    <WarningOutlined />
                </Avatar>
            case 'error':
                return <Avatar
                    sx={{
                        color: 'error.main',
                        bgcolor: 'error.lighter'
                    }}
                >
                    <ExclamationCircleOutlined />
                </Avatar>
            default:
                
            return <Avatar
                sx={{
                    color: 'info.main',
                    bgcolor: 'info.lighter'
                }}
            >
                <InfoCircleOutlined />
            </Avatar>
        }
    };

    const iconBackColor = theme.palette.mode === 'dark' ? 'grey.700' : 'grey.100';
    const iconBackColorOpen = theme.palette.mode === 'dark' ? 'grey.800' : 'grey.200';

    const nbUnread = notifications.filter((notif) => !notif.Read).length;

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
                <Badge badgeContent={nbUnread} color="error">
                    <BellOutlined />
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
                                minWidth: 285,
                                maxWidth: 420,
                                [theme.breakpoints.down('md')]: {
                                    maxWidth: 285
                                }
                            }}
                        >
                            <ClickAwayListener onClickAway={handleClose}>
                                <MainCard
                                    title={t('header.notificationTitle')}
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
                                        {notifications && notifications.map(notification => (<>
                                            <ListItemButton onClick={() => {
                                                notification.Link && redirectToLocal(notification.Link);
                                            }}
                                            style={{
                                                borderLeft: notification.Read ? 'none' : `4px solid ${notification.Level === 'warn' ? theme.palette.warning.main : notification.Level === 'error' ? theme.palette.error.main : theme.palette.info.main}`,
                                                paddingLeft: notification.Read ? '14px' : '10px',
                                            }}>
                                            
                                                <ListItemAvatar>
                                                    {getNotifIcon(notification)}
                                                </ListItemAvatar>
                                                <ListItemText
                                                    primary={<>
                                                        <Typography variant={notification.Read ? 'body' : 'h6'} noWrap>
                                                            {t(notification.Title)}
                                                        </Typography>
                                                        <div style={{ 
                                                            overflow: 'hidden',
                                                            maxHeight: '48px',
                                                            borderLeft: '1px solid grey',
                                                            paddingLeft: '8px',
                                                            margin: '2px'
                                                        }}>
                                                            {t(notification.Message, { Vars: notification.Vars })}
                                                        </div></>
                                                    }
                                                />
                                                <ListItemSecondaryAction>
                                                    <Typography variant="caption" noWrap>
                                                        {format(notification.Date, i18n.resolvedLanguage)}
                                                    </Typography>
                                                </ListItemSecondaryAction>
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

export default Notification;
