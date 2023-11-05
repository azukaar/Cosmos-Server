import { useEffect, useRef, useState } from 'react';

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
    Alert,
    LinearProgress,
    CircularProgress
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
import TableComponent from './components/table';
import { HomeBackground, TransparentHeader } from '../home';
import { formatDate } from './components/utils';
import MiniPlotComponent from './components/mini-plot';
import ResourceDashboard from './resourceDashboard';
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import ProxyDashboard from './proxyDashboard';
import AlertPage from './AlertPage';

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
    const [slot, setSlot] = useState('latest');
    const [currentTab, setCurrentTab] = useState(0);
    const currentTabRef = useRef(currentTab);
    const [coStatus, setCoStatus] = useState(null);
    const [metrics, setMetrics] = useState(null);
    const [isCreatingDB, setIsCreatingDB] = useState(false);

    const [zoom, setZoom] = useState({
        xaxis: {}
    });


    const resetZoom = () => {
        setZoom({
            xaxis: {}
        });
    }

    const refreshMetrics = (override) => {
        let todo = [
            ["cosmos.system.*"],
            ["cosmos.proxy.*"],
            [],
            [],
        ]

        let t = typeof override === 'number' ? override : currentTabRef.current;

        if (t < 2)
            API.metrics.get(todo[t]).then((res) => {
                setMetrics(prevMetrics => {
                    let finalMetrics = prevMetrics ? { ...prevMetrics } : {};
                    if(res.data) {
                        res.data.forEach((metric) => {
                            finalMetrics[metric.Key] = metric;
                        });
                        
                        return finalMetrics;
                    }
                });
            });
    };

    const refreshStatus = () => {
        API.getStatus().then((res) => {
            setCoStatus(res.data);
        });
    }

    useEffect(() => {
        refreshStatus();
        
        let interval = setInterval(() => {
            refreshMetrics();
        }, 10000);

        refreshMetrics();

        return () => {
            clearInterval(interval);
        };
    }, []);
    
    useEffect(() => {
        currentTabRef.current = currentTab;
    }, [currentTab]);

    let xAxis = [];

    if(slot === 'latest') {
      for(let i = 0; i < 100; i++) {
        xAxis.unshift(i);
      }
    }
    else if(slot === 'hourly') {
      for(let i = 0; i < 48; i++) {
        let now = new Date();
        now.setHours(now.getHours() - i);
        now.setMinutes(0);
        now.setSeconds(0);
        xAxis.unshift(formatDate(now, true));
      }
    } else if(slot === 'daily') {
      for(let i = 0; i < 30; i++) {
        let now = new Date();
        now.setDate(now.getDate() - i);
        xAxis.unshift(formatDate(now));
      }
    }

    return (
        <>
        {/* <HomeBackground status={coStatus} />
        <TransparentHeader /> */}
        <IsLoggedIn />
        {!metrics && <Box style={{
          width: '100%',
          height: '100%',
          zIndex: 1000,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginTop: '150px',
        }}>
          <CircularProgress
            size={100}
          />
        </Box>}
        {metrics && <div style={{zIndex:2, position: 'relative'}}>
            <Grid container rowSpacing={4.5} columnSpacing={2.75} >
                <Grid item xs={12} sx={{ mb: -2.25 }}>
                    <Typography variant="h4">Server Monitoring</Typography>
                    {currentTab <= 2 && <Stack direction="row" alignItems="center" spacing={0} style={{marginTop: 10}}>
                        <Button
                            size="small"
                            onClick={() => {setSlot('latest'); resetZoom()}}
                            color={slot === 'latest' ? 'primary' : 'secondary'}
                            variant={slot === 'latest' ? 'outlined' : 'text'}
                        >
                            Latest
                        </Button>
                        <Button
                            size="small"
                            onClick={() => {setSlot('hourly'); resetZoom()}}
                            color={slot === 'hourly' ? 'primary' : 'secondary'}
                            variant={slot === 'hourly' ? 'outlined' : 'text'}
                        >
                            Hourly
                        </Button>
                        <Button
                            size="small"
                            onClick={() => {setSlot('daily'); resetZoom()}}
                            color={slot === 'daily' ? 'primary' : 'secondary'}
                            variant={slot === 'daily' ? 'outlined' : 'text'}
                        >
                            Daily
                        </Button>

                        {zoom.xaxis.min && <Button
                            size="small"
                            onClick={() => {
                                setZoom({
                                    xaxis: {}
                                });
                            }}
                            color={'primary'}
                            variant={'outlined'}
                        >
                            Reset Zoom
                        </Button>}
                    </Stack>}
                    {currentTab > 2 && <div style={{height: 41}}></div>}
                </Grid>


                <Grid item xs={12} md={12} lg={12}>
                    <PrettyTabbedView 
                        currentTab={currentTab}
                        setCurrentTab={(tab) => {
                            setCurrentTab(tab)
                            refreshMetrics(tab);
                        }}
                        fullwidth
                        isLoading={!metrics}
                        tabs={[
                            {
                                title: 'Resources',
                                children: <ResourceDashboard xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} metrics={metrics} />
                            },
                            {
                                title: 'Proxy',
                                children: <ProxyDashboard xAxis={xAxis} zoom={zoom} setZoom={setZoom} slot={slot} metrics={metrics} />
                            },
                            {
                                title: 'Events',
                                children: <AlertPage />
                            },
                            {
                                title: 'Alerts',
                                children: <AlertPage />
                            },
                        ]}
                    />
                </Grid>


                {/* 
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
                */}
                {/* 
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
                 */}
            </Grid>
        </div>}
      </>
    );
};

export default DashboardDefault;
