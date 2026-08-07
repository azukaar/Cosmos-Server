// assets
import { ProfileOutlined, FolderOutlined, PicLeftOutlined, SettingOutlined, NodeExpandOutlined, AppstoreOutlined, ClockCircleOutlined, CloudServerOutlined} from '@ant-design/icons';
import { PERM_RESOURCES_READ, PERM_USERS_READ, PERM_CONFIGURATION_READ } from '../utils/permissions';
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
            url: '/cosmos-ui/servapps',
            icon: AppstoreOutlined,
            permission: PERM_RESOURCES_READ
        },
        {
            id: 'backups',
            title: 'menu-items.management.backups',
            type: 'item',
            url: '/cosmos-ui/backups',
            icon: CloudServerOutlined,
            permission: PERM_RESOURCES_READ
        },
        {
            id: 'url',
            title: 'menu-items.management.urls',
            type: 'item',
            url: '/cosmos-ui/config-url',
            icon: icons.NodeExpandOutlined,
        },
        {
            id: 'storage',
            title: 'menu-items.management.storage',
            type: 'item',
            url: '/cosmos-ui/storage',
            icon: icons.FolderOutlined,
            permission: PERM_RESOURCES_READ
        },
        {
            id: 'constellation',
            title: 'menu-items.management.constellation',
            type: 'item',
            url: '/cosmos-ui/constellation',
            icon: () => <DarkModeSwitch
                light={<img height="28px" width="28px" style={{marginLeft: "-6px"}} src={ConstellationIcon} />}
                dark={<img height="28px" width="28px" style={{marginLeft: "-6px"}} src={ConstellationWhiteIcon} />}
            />,
            
        },
        {
            id: 'users',
            title: 'menu-items.management.usersTitle',
            type: 'item',
            url: '/cosmos-ui/config-users',
            icon: icons.ProfileOutlined,
            permission: PERM_USERS_READ
        },
        {
            id: 'openid',
            title: 'menu-items.management.openId',
            type: 'item',
            url: '/cosmos-ui/openid-manage',
            icon: PicLeftOutlined,
            permission: PERM_CONFIGURATION_READ
        },
        {
            id: 'cron',
            title: 'menu-items.management.schedulerTitle',
            type: 'item',
            url: '/cosmos-ui/cron',
            icon: ClockCircleOutlined,
            permission: PERM_RESOURCES_READ
        },
        {
            id: 'config',
            title: 'menu-items.management.configurationTitle',
            type: 'item',
            url: '/cosmos-ui/config-general',
            icon: icons.SettingOutlined,
        }
    ]
};

export default pages;
