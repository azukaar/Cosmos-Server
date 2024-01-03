import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import MainLayout from '../layout/MainLayout';
import logo from '../assets/images/icons/cosmos.png';
import PrivateRoute from '../PrivateRoute';
import { Navigate } from 'react-router';

const UserManagement = Loadable(lazy(() => import('../pages/config/users/usermanagement')));
const ConfigManagement = Loadable(lazy(() => import('../pages/config/users/configman')));
const ProxyManagement = Loadable(lazy(() => import('../pages/config/users/proxyman')));
const ServAppsIndex = Loadable(lazy(() => import('../pages/servapps/')));
const RouteConfigPage = Loadable(lazy(() => import('../pages/config/routeConfigPage')));
const HomePage = Loadable(lazy(() => import('../pages/home')));
const ContainerIndex = Loadable(lazy(() => import('../pages/servapps/containers')));
const NewDockerServiceForm = Loadable(lazy(() => import('../pages/servapps/containers/newServiceForm')));
const OpenIdList = Loadable(lazy(() => import('../pages/openid/openid-list')));
const MarketPage = Loadable(lazy(() => import('../pages/market/listing')));
const ConstellationIndex = Loadable(lazy(() => import('../pages/constellation')));


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
            element: <RouteConfigPage />
        },
        {
            path: '/cosmos-ui/servapps/containers/:containerName',
            element: <ContainerIndex />
        },
        {
            path: '/cosmos-ui/openid-manage',
            element: <OpenIdList />
        },
        {
            path: '/cosmos-ui/market-listing/',
            element: <MarketPage />
        },
        {
            path: '/cosmos-ui/market-listing/:appStore/:appName',
            element: <MarketPage />
        }
    ].map(children => ({
        ...children,
        element: PrivateRoute({ children: children.element })
    }))
};

export default MainRoutes;
