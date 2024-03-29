// assets
import { ProfileOutlined, PicLeftOutlined, SettingOutlined, NodeExpandOutlined, AppstoreOutlined} from '@ant-design/icons';
import ConstellationIcon from '../assets/images/icons/constellation.png'

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
            url: '/cosmos-ui/servapps',
            icon: AppstoreOutlined,
            adminOnly: true
        },
        {
            id: 'url',
            title: 'URLs',
            type: 'item',
            url: '/cosmos-ui/config-url',
            icon: icons.NodeExpandOutlined,
        },
        {
            id: 'constellation',
            title: 'Constellation',
            type: 'item',
            url: '/cosmos-ui/constellation',
            icon: () => <img height="28px" width="28px" style={{marginLeft: "-6px"}} src={ConstellationIcon} />,
            
        },
        {
            id: 'users',
            title: 'Users',
            type: 'item',
            url: '/cosmos-ui/config-users',
            icon: icons.ProfileOutlined,
            adminOnly: true
        },
        {
            id: 'openid',
            title: 'OpenID',
            type: 'item',
            url: '/cosmos-ui/openid-manage',
            icon: PicLeftOutlined,
            adminOnly: true
        },
        {
            id: 'config',
            title: 'Configuration',
            type: 'item',
            url: '/cosmos-ui/config-general',
            icon: icons.SettingOutlined,
        }
    ]
};

export default pages;
