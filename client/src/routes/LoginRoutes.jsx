import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import { NewMFA, MFALogin } from '../pages/authentication/newMFA';

import MinimalLayout from '../layout/MinimalLayout';
import Logout from '../pages/authentication/Logoff';
import NewInstall from '../pages/newInstall/newInstall';

import ForgotPassword from '../pages/authentication/forgotPassword';
import OpenID from '../pages/authentication/openid';

import AuthLogin from '../pages/authentication/Login';
import AuthRegister from '../pages/authentication/Register';

// ==============================|| AUTH ROUTING ||============================== //

const LoginRoutes = {
    path: '/',
    element: <MinimalLayout />,
    children: [
        {
            path: '/resios-ui/login',
            element: <AuthLogin />
        },
        {
            path: '/resios-ui/register',
            element: <AuthRegister />
        },
        {
            path: '/resios-ui/logout',
            element: <Logout />
        },
        {
            path: '/resios-ui/newInstall',
            element: <NewInstall />
        },
        {
            path: '/resios-ui/newmfa',
            element: <NewMFA />
        },
        {
            path: '/resios-ui/openid',
            element: <OpenID />
        },
        {
            path: '/resios-ui/loginmfa',
            element: <MFALogin />
        },
        {
            path: '/resios-ui/forgot-password',
            element: <ForgotPassword />
        },
    ]
};

export default LoginRoutes;
