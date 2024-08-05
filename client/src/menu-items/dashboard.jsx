// assets
import { HomeOutlined, AppstoreOutlined, DashboardOutlined, AppstoreAddOutlined } from '@ant-design/icons';

// icons
const icons = {
    HomeOutlined
};

// ==============================|| MENU ITEMS - DASHBOARD ||============================== //

const dashboard = {
    id: 'group-dashboard',
    title: 'menu-items.navigation',
    type: 'group',
    children: [
        {
            id: 'home',
            title: 'menu-items.navigation.home',
            type: 'item',
            url: '/cosmos-ui/',
            icon: icons.HomeOutlined,
            breadcrumbs: false
        },
        {
            id: 'dashboard',
            title: 'menu-items.navigation.monitoringTitle',
            type: 'item',
            url: '/cosmos-ui/monitoring',
            icon: DashboardOutlined,
            breadcrumbs: false,
            adminOnly: true
        },
        {
            id: 'market',
            title: 'menu-items.navigation.marketTitle',
            type: 'item',
            url: '/cosmos-ui/market-listing',
            icon: AppstoreAddOutlined,
            breadcrumbs: false
        },
    ]
};

export default dashboard;
