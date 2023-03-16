import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import MainLayout from '../layout/MainLayout';
import UserManagement from '../pages/config/users/usermanagement';
import ConfigManagement from '../pages/config/users/configman';
import ProxyManagement from '../pages/config/users/proxyman';

// render - dashboard
const DashboardDefault = Loadable(lazy(() => import('../pages/dashboard')));

// render - sample page
const SamplePage = Loadable(lazy(() => import('../pages/extra-pages/SamplePage')));

// render - utilities
const Typography = Loadable(lazy(() => import('../pages/components-overview/Typography')));
const Color = Loadable(lazy(() => import('../pages/components-overview/Color')));
const Shadow = Loadable(lazy(() => import('../pages/components-overview/Shadow')));
const AntIcons = Loadable(lazy(() => import('../pages/components-overview/AntIcons')));

// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
    path: '/',
    element: <MainLayout />,
    children: [
        {
            path: '/ui/',
            element: <DashboardDefault />
        },
        {
            path: '/ui/color',
            element: <Color />
        },
        {
            path: '/ui/dashboard',
            children: [
                {
                    path: 'default',
                    element: <DashboardDefault />
                }
            ]
        },
        {
            path: '/ui/config/users',
            element: <UserManagement />
        },
        {
            path: '/ui/config/general',
            element: <ConfigManagement />
        },
        {
            path: '/ui/config/proxy',
            element: <ProxyManagement />
        },
        {
            path: '/ui/shadow',
            element: <Shadow />
        },
        {
            path: '/ui/typography',
            element: <Typography />
        },
        {
            path: '/ui/icons/ant',
            element: <AntIcons />
        }
    ]
};

export default MainRoutes;
