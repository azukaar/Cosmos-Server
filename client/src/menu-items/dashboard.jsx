// assets
import { HomeOutlined, AppstoreOutlined, DashboardOutlined } from '@ant-design/icons';

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
    ]
};

export default dashboard;
