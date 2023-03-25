// assets
import { HomeOutlined, AppstoreOutlined } from '@ant-design/icons';

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
            url: '/ui',
            icon: icons.HomeOutlined,
            breadcrumbs: false
        },
        {
            id: 'servapps',
            title: 'ServApps',
            type: 'item',
            url: '/ui/servapps',
            icon: AppstoreOutlined
        },
    ]
};

export default dashboard;
