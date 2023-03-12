
import * as API from './api';
import { useEffect } from 'react';

const isLoggedIn = () => useEffect(() => {
    API.auth.me().then((data) => {
        if(data.status != 'OK') {
            window.location.href = '/login';
        }
    });
});

export default isLoggedIn;