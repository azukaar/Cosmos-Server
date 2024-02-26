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
