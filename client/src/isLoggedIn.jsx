
import * as API from './api';
import { useEffect } from 'react';

const isLoggedIn = () => useEffect(() => {
    console.log("CHECK LOGIN")
    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            window.location.href = '/ui/login';
        }
    });
}, []);

export default isLoggedIn;