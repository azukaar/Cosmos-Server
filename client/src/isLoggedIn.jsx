
import * as API from './api';
import { useEffect } from 'react';

const isLoggedIn = () => useEffect(() => {
    console.log("CHECK LOGIN")
    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            if(data.status == 'NEW_INSTALL') {
                window.location.href = '/ui/newInstall';
            } else
                window.location.href = '/ui/login';
        }
    });
}, []);

export default isLoggedIn;