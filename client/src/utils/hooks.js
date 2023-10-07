import React from 'react';
import { useCookies } from 'react-cookie';
import { logout } from '../api/authentication';

function useClientInfos() {
  const [cookies] = useCookies(['client-infos', 'jwttoken']);

  if(cookies['jwttoken'] === undefined) {
    // probably the demo or new install
    return {
      nickname: "",
      role: "2"
    }
  }
  
  let clientInfos = null;
  
  try {
    // Try to parse the cookie into a JavaScript object
    clientInfos = cookies['client-infos'].split(',');
    
    return {
      nickname: clientInfos[0],
      role: clientInfos[1]
    };
  } catch (error) {
    console.error('Error parsing client-infos cookie:', error);
    logout();
  }
}

export {
  useClientInfos
};