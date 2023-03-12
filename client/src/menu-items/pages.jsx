// assets
import { ProfileOutlined, SettingOutlined, NodeExpandOutlined} from '@ant-design/icons';

// icons
const icons = {
    NodeExpandOutlined,
    ProfileOutlined,
    SettingOutlined
};

// ==============================|| MENU ITEMS - EXTRA PAGES ||============================== //

const pages = {
    id: 'management',
    title: 'Management',
    type: 'group',
    children: [
        {
            id: 'proxy',
            title: 'Proxy Routes',
            type: 'item',
            url: '/config/proxy',
            icon: icons.NodeExpandOutlined,
        },
        {
            id: 'users',
            title: 'Manage Users',
            type: 'item',
            url: '/config/users',
            icon: icons.ProfileOutlined,
        },
        {
            id: 'config',
            title: 'Configuration',
            type: 'item',
            url: '/config/general',
            icon: icons.SettingOutlined,
        }
    ]
};

export default pages;
