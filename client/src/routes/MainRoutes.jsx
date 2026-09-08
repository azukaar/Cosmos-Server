// project import
import MainLayout from '../layout/MainLayout';
import logo from '../assets/images/icons/logo2.png';
import { Navigate } from 'react-router';
import UserManagement from '../pages/config/users/usermanagement';
import ConfigManagement from '../pages/config';
import ProxyManagement from '../pages/config/users/proxyman';
import ServAppsIndex from '../pages/servapps/';
import RouteConfigPage from '../pages/config/routeConfigPage';
import HomePage from '../pages/home';
import ContainerIndex from '../pages/servapps/containers';
import NewDockerServiceForm from '../pages/servapps/containers/newServiceForm';
import OpenIdList from '../pages/openid/openid-list';
import MarketPage from '../pages/market/listing';
import ConstellationIndex  from '../pages/constellation';
import proFeatures from '../pro';
import { ProRoute } from '../components/proPage';
import StorageIndex from '../pages/storage';
import DashboardDefault from '../pages/dashboard';
import { CronManager } from '../pages/cron/jobsManage';
import PrivateRoute from '../PrivateRoute';
import TrustPage from '../pages/config/trust';
import AllBackupsIndex from '../pages/backups';
import SingleBackupIndex from '../pages/backups/single-backup-index';
import SingleRepoIndex from '../pages/backups/single-repo-index';
import { isDomain } from '../utils/indexs';

// ==============================|| MAIN ROUTING ||============================== //

const MainRoutes = {
    path: '/',
    element: <MainLayout />,
    children: [
        {
            path: '/',
            // redirect to /cosmos-ui
            element: <Navigate to="/cosmos-ui/" />
        },
        {
            path: '/cosmos-ui/logo',
            // redirect to /cosmos-ui
            element: <Navigate to={logo} />
        },
        [{
            path: '/cosmos-ui',
            element: <HomePage />
        },
        {
            path: '/cosmos-ui/monitoring',
            element: <DashboardDefault />
        },
        {
            path: '/cosmos-ui/backups',
            element: <AllBackupsIndex />
        },
        {
            path: '/cosmos-ui/backups/:backupName/',
            element: <SingleBackupIndex />
        },
        {
            path: '/cosmos-ui/backups/repo/:backupName/',
            element: <SingleRepoIndex />
        },
        {
            path: '/cosmos-ui/backups/:backupName/:subpath',
            element: <SingleBackupIndex />
        },
        {
            path: '/cosmos-ui/storage',
            element: <StorageIndex />
        },
        {
            path: '/cosmos-ui/constellation',
            element: <ConstellationIndex />
        },
        {
            path: '/cosmos-ui/constellation/*',
            element: <ConstellationIndex />
        },
        {
            path: '/cosmos-ui/deployments',
            element: <ProRoute component={proFeatures.DeploymentsPage} />
        },
        {
            path: '/cosmos-ui/deployments/view/:name',
            element: <ProRoute component={proFeatures.DeploymentPage} />
        },
        {
            path: '/cosmos-ui/deployments/*',
            element: <ProRoute component={proFeatures.DeploymentsPage} />
        },
        {
            path: '/cosmos-ui/functions',
            element: <ProRoute component={proFeatures.FunctionsPage} />
        },
        {
            path: '/cosmos-ui/functions/view/:name',
            element: <ProRoute component={proFeatures.FunctionPage} />
        },
        {
            path: '/cosmos-ui/functions/*',
            element: <ProRoute component={proFeatures.FunctionsPage} />
        },
        {
            path: '/cosmos-ui/databases',
            element: <ProRoute component={proFeatures.DatabasesPage} />
        },
        {
            path: '/cosmos-ui/databases/view/:name',
            element: <ProRoute component={proFeatures.DatabasePage} />
        },
        {
            path: '/cosmos-ui/databases/*',
            element: <ProRoute component={proFeatures.DatabasesPage} />
        },
        {
            path: '/cosmos-ui/object-storage',
            element: <ProRoute component={proFeatures.ObjectStoragePage} />
        },
        {
            path: '/cosmos-ui/object-storage/view/:name',
            element: <ProRoute component={proFeatures.SeaweedFSInstancePage} />
        },
        {
            path: '/cosmos-ui/object-storage/*',
            element: <ProRoute component={proFeatures.ObjectStoragePage} />
        },
        {
            path: '/cosmos-ui/registries',
            element: <ProRoute component={proFeatures.RegistriesPage} />
        },
        {
            path: '/cosmos-ui/registries/view/:name',
            element: <ProRoute component={proFeatures.RegistryPage} />
        },
        {
            path: '/cosmos-ui/registries/access/:name',
            element: <ProRoute component={proFeatures.RegistryAccessPage} />
        },
        {
            path: '/cosmos-ui/registries/*',
            element: <ProRoute component={proFeatures.RegistriesPage} />
        },
        {
            path: '/cosmos-ui/trust',
            element: <TrustPage />
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
            path: '/cosmos-ui/config-users/*',
            element: <UserManagement />
        },
        {
            path: '/cosmos-ui/config-general',
            element: <ConfigManagement />
        },
        {
            path: '/cosmos-ui/config-general/*',
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
        },
        {
            path: '/cosmos-ui/cron',
            element: <CronManager />
        }].map(children => ({
            ...children,
            element: PrivateRoute({ children: children.element })
        }))
    ].flat()
};

export default MainRoutes;
