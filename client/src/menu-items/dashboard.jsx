// assets
import { HomeOutlined, AppstoreOutlined, DashboardOutlined, AppstoreAddOutlined } from '@ant-design/icons';

// icons
const icons = {
    HomeOutlined
};

// ==============================|| MENU ITEMS - DASHBOARD ||============================== //

const dashboard = {
    id: 'group-dashboard',
    title: 'Navigation',
    type: 'group',
    children: [
        {
            id: 'home',
            title: 'Home',
            type: 'item',
            url: '/ui/',
            icon: icons.HomeOutlined,
            breadcrumbs: false
        },
        {
            id: 'dashboard',
            title: 'Dashboard',
            type: 'item',
            url: '/ui/dashboard',
            icon: DashboardOutlined,
            breadcrumbs: false
        },
        {
            id: 'market',
            title: 'Market',
            type: 'item',
            url: '/ui/market-listing',
            icon: AppstoreAddOutlined,
            breadcrumbs: false
        },
    ]
};

export default dashboard;
