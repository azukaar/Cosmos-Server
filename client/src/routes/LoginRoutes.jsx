import { lazy } from 'react';
import Loadable from '../components/Loadable';

// project import

import { NewMFA, MFALogin } from '../pages/authentication/newMFA';

const MinimalLayout = Loadable(lazy(() => import('../layout/MinimalLayout')));
const Logout = Loadable(lazy(() => import('../pages/authentication/Logoff')));
const NewInstall = Loadable(lazy(() => import('../pages/newInstall/newInstall')))

const ForgotPassword = Loadable(lazy(() => import('../pages/authentication/forgotPassword')));
const OpenID = Loadable(lazy(() => import('../pages/authentication/openid')));

// render - login
const AuthLogin = Loadable(lazy(() => import('../pages/authentication/Login')));
const AuthRegister = Loadable(lazy(() => import('../pages/authentication/Register')));

// ==============================|| AUTH ROUTING ||============================== //

const LoginRoutes = {
    path: '/',
    element: <MinimalLayout />,
    children: [
        {
            path: '/cosmos-ui/login',
            element: <AuthLogin />
        },
        {
            path: '/cosmos-ui/register',
            element: <AuthRegister />
        },
        {
            path: '/cosmos-ui/logout',
            element: <Logout />
        },
        {
            path: '/cosmos-ui/newInstall',
            element: <NewInstall />
        },
        {
            path: '/cosmos-ui/newmfa',
            element: <NewMFA />
        },
        {
            path: '/cosmos-ui/openid',
            element: <OpenID />
        },
        {
            path: '/cosmos-ui/loginmfa',
            element: <MFALogin />
        },
        {
            path: '/cosmos-ui/forgot-password',
            element: <ForgotPassword />
        },
    ]
};

export default LoginRoutes;
