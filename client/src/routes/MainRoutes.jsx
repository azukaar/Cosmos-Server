import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import MainLayout from '../layout/MainLayout';
import UserManagement from '../pages/config/users/usermanagement';
import ConfigManagement from '../pages/config/users/configman';
import ProxyManagement from '../pages/config/users/proxyman';
import ServAppsIndex from '../pages/servapps/';
import { Navigate } from 'react-router';
import RouteConfigPage from '../pages/config/routeConfigPage';
import logo from '../assets/images/icons/cosmos.png';
import HomePage from '../pages/home';
import ContainerIndex from '../pages/servapps/containers';
import NewDockerServiceForm from '../pages/servapps/containers/newServiceForm';
import OpenIdList from '../pages/openid/openid-list';
import MarketPage from '../pages/market/listing';
import ConstellationIndex  from '../pages/constellation';


// render - dashboard
const DashboardDefault = Loadable(lazy(() => import('../pages/dashboard')));

// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
    path: '/',
    element: <MainLayout />,
    children: [
        {
            path: '/',
            // redirect to /cosmos-ui
            element: <Navigate to="/cosmos-ui" />
        },
        {
            path: '/cosmos-ui/logo',
            // redirect to /cosmos-ui
            element: <Navigate to={logo} />
        },
        {
            path: '/cosmos-ui',
            element: <HomePage />
        },
        {
            path: '/cosmos-ui/monitoring',
            element: <DashboardDefault />
        },
        {
            path: '/cosmos-ui/constellation',
            element: <ConstellationIndex />
        },
        {
            path: '/cosmos-ui/servapps',
            element: <ServAppsIndex />
        },
        {
            path: '/cosmos-ui/servapps/stack/:stack',
            element: <ServAppsIndex />
        },
        {
            path: '/cosmos-ui/config-users',
            element: <UserManagement />
        },
        {
            path: '/cosmos-ui/config-general',
            element: <ConfigManagement />
        },
        {
            path: '/cosmos-ui/servapps/new-service',
            element: <NewDockerServiceForm />
        },
        {
            path: '/cosmos-ui/config-url',
            element: <ProxyManagement />
        },
        {
            path: '/cosmos-ui/config-url/:routeName',
            element: <RouteConfigPage />,
        },
        {
            path: '/cosmos-ui/servapps/containers/:containerName',
            element: <ContainerIndex />,
        },
        {
            path: '/cosmos-ui/openid-manage',
            element: <OpenIdList />,
        },
        {
            path: '/cosmos-ui/market-listing/',
            element: <MarketPage />
        },
        {
            path: '/cosmos-ui/market-listing/:appStore/:appName',
            element: <MarketPage />
        }
    ]
};

export default MainRoutes;
