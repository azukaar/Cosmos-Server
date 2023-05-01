// assets
import { ProfileOutlined, SettingOutlined, NodeExpandOutlined, AppstoreOutlined} from '@ant-design/icons';

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
            id: 'servapps',
            title: 'ServApps',
            type: 'item',
            url: '/ui/servapps',
            icon: AppstoreOutlined
        },
        {
            id: 'url',
            title: 'URLs',
            type: 'item',
            url: '/ui/config-url',
            icon: icons.NodeExpandOutlined,
        },
        {
            id: 'users',
            title: 'Users',
            type: 'item',
            url: '/ui/config-users',
            icon: icons.ProfileOutlined,
        },
        {
            id: 'config',
            title: 'Configuration',
            type: 'item',
            url: '/ui/config-general',
            icon: icons.SettingOutlined,
        }
    ]
};

export default pages;
