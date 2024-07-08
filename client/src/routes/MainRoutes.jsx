// project import
import MainLayout from '../layout/MainLayout';
import logo from '../assets/images/icons/cosmos.png';
import { Navigate } from 'react-router';
import UserManagement from '../pages/config/users/usermanagement';
import ConfigManagement from '../pages/config/users/configman';
import ProxyManagement from '../pages/config/users/proxyman';
import ServAppsIndex from '../pages/servapps/';
import RouteConfigPage from '../pages/config/routeConfigPage';
import HomePage from '../pages/home';
import ContainerIndex from '../pages/servapps/containers';
import NewDockerServiceForm from '../pages/servapps/containers/newServiceForm';
import OpenIdList from '../pages/openid/openid-list';
import MarketPage from '../pages/market/listing';
import ConstellationIndex  from '../pages/constellation';
import StorageIndex from '../pages/storage';
import DashboardDefault from '../pages/dashboard';
import { CronManager } from '../pages/cron/jobsManage';
import PrivateRoute from '../PrivateRoute';

// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
    path: '/',
    element: <MainLayout />,
    children: [
        {
            path: '/',
            // redirect to /resios-ui
            element: <Navigate to="/resios-ui/" />
        },
        {
            path: '/resios-ui/logo',
            // redirect to /resios-ui
            element: <Navigate to={logo} />
        },
        [{
            path: '/resios-ui',
            element: <HomePage />
        },
        {
            path: '/resios-ui/monitoring',
            element: <DashboardDefault />
        },
        {
            path: '/resios-ui/storage',
            element: <StorageIndex />
        },
        {
            path: '/resios-ui/constellation',
            element: <ConstellationIndex />
        },
        {
            path: '/resios-ui/servapps',
            element: <ServAppsIndex />
        },
        {
            path: '/resios-ui/servapps/stack/:stack',
            element: <ServAppsIndex />
        },
        {
            path: '/resios-ui/config-users',
            element: <UserManagement />
        },
        {
            path: '/resios-ui/config-general',
            element: <ConfigManagement />
        },
        {
            path: '/resios-ui/servapps/new-service',
            element: <NewDockerServiceForm />
        },
        {
            path: '/resios-ui/config-url',
            element: <ProxyManagement />
        },
        {
            path: '/resios-ui/config-url/:routeName',
            element: <RouteConfigPage />,
        },
        {
            path: '/resios-ui/servapps/containers/:containerName',
            element: <ContainerIndex />,
        },
        {
            path: '/resios-ui/openid-manage',
            element: <OpenIdList />,
        },
        {
            path: '/resios-ui/market-listing/',
            element: <MarketPage />
        },
        {
            path: '/resios-ui/market-listing/:appStore/:appName',
            element: <MarketPage />
        },
        {
            path: '/resios-ui/cron',
            element: <CronManager />
        }].map(children => ({
            ...children,
            element: PrivateRoute({ children: children.element })
        }))
    ].flat()
};

export default MainRoutes;
