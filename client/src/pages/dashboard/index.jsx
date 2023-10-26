import { useEffect, useState } from 'react';

// material-ui
import {
    Avatar,
    AvatarGroup,
    Box,
    Button,
    Grid,
    List,
    ListItemAvatar,
    ListItemButton,
    ListItemSecondaryAction,
    ListItemText,
    MenuItem,
    Stack,
    TextField,
    Typography,
    Alert
} from '@mui/material';

// project import
import OrdersTable from './OrdersTable';
import IncomeAreaChart from './IncomeAreaChart';
import MonthlyBarChart from './MonthlyBarChart';
import ReportAreaChart from './ReportAreaChart';
import SalesColumnChart from './SalesColumnChart';
import MainCard from '../../components/MainCard';
import AnalyticEcommerce from '../../components/cards/statistics/AnalyticEcommerce';

// assets
import { GiftOutlined, MessageOutlined, SettingOutlined } from '@ant-design/icons';
import avatar1 from '../../assets/images/users/avatar-1.png';
import avatar2 from '../../assets/images/users/avatar-2.png';
import avatar3 from '../../assets/images/users/avatar-3.png';
import avatar4 from '../../assets/images/users/avatar-4.png';
import IsLoggedIn from '../../isLoggedIn';

import * as API from '../../api';
import AnimateButton from '../../components/@extended/AnimateButton';
import PlotComponent from './components/plot';

// avatar style
const avatarSX = {
    width: 36,
    height: 36,
    fontSize: '1rem'
};

// action style
const actionSX = {
    mt: 0.75,
    ml: 1,
    top: 'auto',
    right: 'auto',
    alignSelf: 'flex-start',
    transform: 'none'
};

// sales report status
const status = [
    {
        value: 'today',
        label: 'Today'
    },
    {
        value: 'month',
        label: 'This Month'
    },
    {
        value: 'year',
        label: 'This Year'
    }
];

// ==============================|| DASHBOARD - DEFAULT ||============================== //

const DashboardDefault = () => {
    const [value, setValue] = useState('today');
    const [slot, setSlot] = useState('week');

    const [coStatus, setCoStatus] = useState(null);
    const [metrics, setMetrics] = useState(null);
    const [isCreatingDB, setIsCreatingDB] = useState(false);

    const refreshMetrics = () => {
        API.metrics.get().then((res) => {
            let finalMetrics = {};
            res.data.forEach((metric) => {
                finalMetrics[metric.Key] = metric;
            });
            setMetrics(finalMetrics);
            setTimeout(refreshMetrics, 10000);
        });
    };

    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    useEffect(() => {
        refreshStatus();
        refreshMetrics();
    }, []);

    return (
        <>
        <IsLoggedIn />
        <div>
            <Stack spacing={1}>
                {coStatus && !coStatus.database && (
                    <Alert severity="error">
                        No Database is setup for Cosmos! User Management and Authentication will not work.<br />
                        You can either setup the database, or disable user management in the configuration panel.<br />
                    </Alert>
                )}

                {coStatus && coStatus.letsencrypt && (
                    <Alert severity="error">
                        You have enabled Let's Encrypt for automatic HTTPS Certificate. You need to provide the configuration with an email address to use for Let's Encrypt in the configs.
                    </Alert>
                )}

                {coStatus && coStatus.newVersionAvailable && (
                    <Alert severity="warning">
                        A new version of Cosmos is available! Please update to the latest version to get the latest features and bug fixes.
                    </Alert>
                )}

                {coStatus && coStatus.needsRestart && (
                    <Alert severity="warning">
                        You have made changes to the configuration that require a restart to take effect. Please restart Cosmos to apply the changes.
                    </Alert>
                )}

                {coStatus && coStatus.domain && (
                    <Alert severity="error">
                        You are using localhost or 0.0.0.0 as a hostname in the configuration. It is recommended that you use a domain name instead.
                    </Alert>
                )}

                {coStatus && !coStatus.docker && (
                    <Alert severity="error">
                        Docker is not connected! Please check your docker connection.<br/>
                        Did you forget to add <pre>-v /var/run/docker.sock:/var/run/docker.sock</pre> to your docker run command?<br />
                        if your docker daemon is running somewhere else, please add <pre>-e DOCKER_HOST=...</pre> to your docker run command.
                    </Alert>
                )}

                <Alert severity="info">Dashboard implementation currently in progress! If you want to voice your opinion on where Cosmos is going, please join us on Discord!</Alert>
            </Stack>
        </div>
        {metrics && <div style={{marginTop: '30px'}}>
            <Grid container rowSpacing={4.5} columnSpacing={2.75}>
                {/* row 1 */}
                <Grid item xs={12} sx={{ mb: -2.25 }}>
                    <Typography variant="h5">Dashboard</Typography>
                </Grid>
                <Grid item xs={12} sm={6} md={4} lg={3}>
                    <AnalyticEcommerce title="Total Page Views" count="4,42,236" percentage={59.3} extra="35,000" />
                </Grid>
                <Grid item xs={12} sm={6} md={4} lg={3}>
                    <AnalyticEcommerce title="Total Users" count="78,250" percentage={70.5} extra="8,900" />
                </Grid>
                <Grid item xs={12} sm={6} md={4} lg={3}>
                    <AnalyticEcommerce title="Total Order" count="18,800" percentage={27.4} isLoss color="warning" extra="1,943" />
                </Grid>
                <Grid item xs={12} sm={6} md={4} lg={3}>
                    <AnalyticEcommerce title="Total Sales" count="$35,078" percentage={27.4} isLoss color="warning" extra="$20,395" />
                </Grid>

                <Grid item md={8} sx={{ display: { sm: 'none', md: 'block', lg: 'none' } }} />

                {/* row 2 */}
                
                <PlotComponent data={[metrics["cosmos.system.cpu.0"], metrics["cosmos.system.ram"]]}/>
                
                <Grid item xs={12} md={5} lg={4}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h5">Income Overview</Typography>
                        </Grid>
                        <Grid item />
                    </Grid>
                    <MainCard sx={{ mt: 2 }} content={false}>
                        <Box sx={{ p: 3, pb: 0 }}>
                            <Stack spacing={2}>
                                <Typography variant="h6" color="textSecondary">
                                    This Week Statistics
                                </Typography>
                                <Typography variant="h3">$7,650</Typography>
                            </Stack>
                        </Box>
                        <MonthlyBarChart />
                    </MainCard>
                </Grid>

                {/* row 3 */}
                <Grid item xs={12} md={7} lg={8}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h5">Recent Orders</Typography>
                        </Grid>
                        <Grid item />
                    </Grid>
                    <MainCard sx={{ mt: 2 }} content={false}>
                        <OrdersTable />
                    </MainCard>
                </Grid>
                <Grid item xs={12} md={5} lg={4}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h5">Analytics Report</Typography>
                        </Grid>
                        <Grid item />
                    </Grid>
                    <MainCard sx={{ mt: 2 }} content={false}>
                        <List sx={{ p: 0, '& .MuiListItemButton-root': { py: 2 } }}>
                            <ListItemButton divider>
                                <ListItemText primary="Company Finance Growth" />
                                <Typography variant="h5">+45.14%</Typography>
                            </ListItemButton>
                            <ListItemButton divider>
                                <ListItemText primary="Company Expenses Ratio" />
                                <Typography variant="h5">0.58%</Typography>
                            </ListItemButton>
                            <ListItemButton>
                                <ListItemText primary="Business Risk Cases" />
                                <Typography variant="h5">Low</Typography>
                            </ListItemButton>
                        </List>
                        <ReportAreaChart />
                    </MainCard>
                </Grid>

                {/* row 4 */}
                <Grid item xs={12} md={7} lg={8}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h5">Sales Report</Typography>
                        </Grid>
                        <Grid item>
                            <TextField
                                id="standard-select-currency"
                                size="small"
                                select
                                value={value}
                                onChange={(e) => setValue(e.target.value)}
                                sx={{ '& .MuiInputBase-input': { py: 0.5, fontSize: '0.875rem' } }}
                            >
                                {status.map((option) => (
                                    <MenuItem key={option.value} value={option.value}>
                                        {option.label}
                                    </MenuItem>
                                ))}
                            </TextField>
                        </Grid>
                    </Grid>
                    <MainCard sx={{ mt: 1.75 }}>
                        <Stack spacing={1.5} sx={{ mb: -12 }}>
                            <Typography variant="h6" color="secondary">
                                Net Profit
                            </Typography>
                            <Typography variant="h4">$1560</Typography>
                        </Stack>
                        <SalesColumnChart />
                    </MainCard>
                </Grid>
                <Grid item xs={12} md={5} lg={4}>
                    <Grid container alignItems="center" justifyContent="space-between">
                        <Grid item>
                            <Typography variant="h5">Transaction History</Typography>
                        </Grid>
                        <Grid item />
                    </Grid>
                    <MainCard sx={{ mt: 2 }} content={false}>
                        <List
                            component="nav"
                            sx={{
                                px: 0,
                                py: 0,
                                '& .MuiListItemButton-root': {
                                    py: 1.5,
                                    '& .MuiAvatar-root': avatarSX,
                                    '& .MuiListItemSecondaryAction-root': { ...actionSX, position: 'relative' }
                                }
                            }}
                        >
                            <ListItemButton divider>
                                <ListItemAvatar>
                                    <Avatar
                                        sx={{
                                            color: 'success.main',
                                            bgcolor: 'success.lighter'
                                        }}
                                    >
                                        <GiftOutlined />
                                    </Avatar>
                                </ListItemAvatar>
                                <ListItemText primary={<Typography variant="subtitle1">Order #002434</Typography>} secondary="Today, 2:00 AM" />
                                <ListItemSecondaryAction>
                                    <Stack alignItems="flex-end">
                                        <Typography variant="subtitle1" noWrap>
                                            + $1,430
                                        </Typography>
                                        <Typography variant="h6" color="secondary" noWrap>
                                            78%
                                        </Typography>
                                    </Stack>
                                </ListItemSecondaryAction>
                            </ListItemButton>
                            <ListItemButton divider>
                                <ListItemAvatar>
                                    <Avatar
                                        sx={{
                                            color: 'primary.main',
                                            bgcolor: 'primary.lighter'
                                        }}
                                    >
                                        <MessageOutlined />
                                    </Avatar>
                                </ListItemAvatar>
                                <ListItemText
                                    primary={<Typography variant="subtitle1">Order #984947</Typography>}
                                    secondary="5 August, 1:45 PM"
                                />
                                <ListItemSecondaryAction>
                                    <Stack alignItems="flex-end">
                                        <Typography variant="subtitle1" noWrap>
                                            + $302
                                        </Typography>
                                        <Typography variant="h6" color="secondary" noWrap>
                                            8%
                                        </Typography>
                                    </Stack>
                                </ListItemSecondaryAction>
                            </ListItemButton>
                            <ListItemButton>
                                <ListItemAvatar>
                                    <Avatar
                                        sx={{
                                            color: 'error.main',
                                            bgcolor: 'error.lighter'
                                        }}
                                    >
                                        <SettingOutlined />
                                    </Avatar>
                                </ListItemAvatar>
                                <ListItemText primary={<Typography variant="subtitle1">Order #988784</Typography>} secondary="7 hours ago" />
                                <ListItemSecondaryAction>
                                    <Stack alignItems="flex-end">
                                        <Typography variant="subtitle1" noWrap>
                                            + $682
                                        </Typography>
                                        <Typography variant="h6" color="secondary" noWrap>
                                            16%
                                        </Typography>
                                    </Stack>
                                </ListItemSecondaryAction>
                            </ListItemButton>
                        </List>
                    </MainCard>
                    <MainCard sx={{ mt: 2 }}>
                        <Stack spacing={3}>
                            <Grid container justifyContent="space-between" alignItems="center">
                                <Grid item>
                                    <Stack>
                                        <Typography variant="h5" noWrap>
                                            Help & Support Chat
                                        </Typography>
                                        <Typography variant="caption" color="secondary" noWrap>
                                            Typical replay within 5 min
                                        </Typography>
                                    </Stack>
                                </Grid>
                                <Grid item>
                                    <AvatarGroup sx={{ '& .MuiAvatar-root': { width: 32, height: 32 } }}>
                                        <Avatar alt="Remy Sharp" src={avatar1} />
                                        <Avatar alt="Travis Howard" src={avatar2} />
                                        <Avatar alt="Cindy Baker" src={avatar3} />
                                        <Avatar alt="Agnes Walker" src={avatar4} />
                                    </AvatarGroup>
                                </Grid>
                            </Grid>
                            <Button size="small" variant="contained" sx={{ textTransform: 'capitalize' }}>
                                Need Help?
                            </Button>
                        </Stack>
                    </MainCard>
                </Grid>
            </Grid>
        </div>}
      </>
    );
};

export default DashboardDefault;
