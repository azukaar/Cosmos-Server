import path from 'path';
import { lazy } from 'react';

// project import
import Loadable from '../components/Loadable';
import MinimalLayout from '../layout/MinimalLayout';
import Logout from '../pages/authentication/Logoff';
import NewInstall from '../pages/newInstall/newInstall';

import {NewMFA, MFALogin} from '../pages/authentication/newMFA';

// render - login
const AuthLogin = Loadable(lazy(() => import('../pages/authentication/Login')));
const AuthRegister = Loadable(lazy(() => import('../pages/authentication/Register')));

// ==============================|| AUTH ROUTING ||============================== //

const LoginRoutes = {
    path: '/',
    element: <MinimalLayout />,
    children: [
        {
            path: '/ui/login',
            element: <AuthLogin />
        },
        {
            path: '/ui/register',
            element: <AuthRegister />
        },
        {
            path: '/ui/logout',
            element: <Logout />
        },
        {
            path: '/ui/newInstall',
            element: <NewInstall />
        },
        {
            path: '/ui/newmfa',
            element: <NewMFA />
        },
        {
            path: '/ui/loginmfa',
            element: <MFALogin />
        },
    ]
};

export default LoginRoutes;
