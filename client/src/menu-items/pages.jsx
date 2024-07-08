// assets
import { ProfileOutlined, FolderOutlined, PicLeftOutlined, SettingOutlined, NodeExpandOutlined, AppstoreOutlined, ClockCircleOutlined} from '@ant-design/icons';
import ConstellationIcon from '../assets/images/icons/constellation.png';
import ConstellationWhiteIcon from '../assets/images/icons/constellation_white.png';
import { DarkModeSwitch } from '../utils/indexs';

// icons
const icons = {
    NodeExpandOutlined,
    ProfileOutlined,
    SettingOutlined,
    FolderOutlined
};
// ==============================|| MENU ITEMS - EXTRA PAGES ||============================== //

const pages = {
    id: 'management',
    title: 'menu-items.managementTitle',
    type: 'group',
    children: [
        {
            id: 'servapps',
            title: 'menu-items.management.servApps',
            type: 'item',
            url: '/resios-ui/servapps',
            icon: AppstoreOutlined,
            adminOnly: true
        },
        {
            id: 'url',
            title: 'menu-items.management.urls',
            type: 'item',
            url: '/resios-ui/config-url',
            icon: icons.NodeExpandOutlined,
        },
        {
            id: 'users',
            title: 'menu-items.management.usersTitle',
            type: 'item',
            url: '/resios-ui/config-users',
            icon: icons.ProfileOutlined,
            adminOnly: true
        },
        {
            id: 'openid',
            title: 'menu-items.management.openId',
            type: 'item',
            url: '/resios-ui/openid-manage',
            icon: PicLeftOutlined,
            adminOnly: true
        },
        {
            id: 'cron',
            title: 'menu-items.management.schedulerTitle',
            type: 'item',
            url: '/resios-ui/cron',
            icon: ClockCircleOutlined,
            adminOnly: true
        },
        {
            id: 'config',
            title: 'menu-items.management.configurationTitle',
            type: 'item',
            url: '/resios-ui/config-general',
            icon: icons.SettingOutlined,
        }
    ]
};

export default pages;
