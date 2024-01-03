import path from 'path';
import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import MinimalLayout from '../layout/MinimalLayout';
import Logout from '../pages/authentication/Logoff';
const NewInstall = Loadable(lazy(() => import('../pages/newInstall/newInstall')))

import {NewMFA, MFALogin} from '../pages/authentication/newMFA';
import ForgotPassword from '../pages/authentication/forgotPassword';
import OpenID from '../pages/authentication/openid';

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
