// assets
import { CodeSandboxOutlined, DatabaseOutlined, DeploymentUnitOutlined, FunctionOutlined, InboxOutlined } from '@ant-design/icons';
import { PERM_RESOURCES_READ } from '../utils/permissions';
import ConstellationIcon from '../assets/images/icons/constellation.png';
import ConstellationWhiteIcon from '../assets/images/icons/constellation_white.png';
import { DarkModeSwitch } from '../utils/indexs';

// ==============================|| MENU ITEMS - CONSTELLATION ||============================== //

const constellation = {
    id: 'group-constellation',
    title: 'menu-items.constellationTitle',
    type: 'group',
    children: [
        {
            id: 'constellation',
            title: 'menu-items.constellation.network',
            type: 'item',
            url: '/cosmos-ui/constellation',
            icon: () => <DarkModeSwitch
                light={<img height="28px" width="28px" style={{marginLeft: "-6px"}} src={ConstellationIcon} />}
                dark={<img height="28px" width="28px" style={{marginLeft: "-6px"}} src={ConstellationWhiteIcon} />}
            />,
        },
        {
            id: 'deployments',
            title: 'menu-items.constellation.deployments',
            type: 'item',
            url: '/cosmos-ui/deployments',
            icon: DeploymentUnitOutlined,
            permission: PERM_RESOURCES_READ,
        },
        {
            id: 'databases',
            title: 'menu-items.constellation.databases',
            type: 'item',
            url: '/cosmos-ui/databases',
            icon: DatabaseOutlined,
            permission: PERM_RESOURCES_READ,
        },
        {
            id: 'object-storage',
            title: 'menu-items.constellation.objectStorage',
            type: 'item',
            url: '/cosmos-ui/object-storage',
            icon: InboxOutlined,
            permission: PERM_RESOURCES_READ,
        },
        {
            id: 'registries',
            title: 'menu-items.constellation.registries',
            type: 'item',
            url: '/cosmos-ui/registries',
            icon: CodeSandboxOutlined,
            permission: PERM_RESOURCES_READ,
        },
        {
            id: 'functions',
            title: 'menu-items.constellation.functions',
            type: 'item',
            url: '/cosmos-ui/functions',
            icon: FunctionOutlined,
            permission: PERM_RESOURCES_READ,
        },
    ]
};

export default constellation;
