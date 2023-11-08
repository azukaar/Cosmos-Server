import React from 'react';
import { useCookies } from 'react-cookie';
import { logout } from '../api/authentication';

const isDemo = import.meta.env.MODE === 'demo';

function useClientInfos() {
  const [cookies] = useCookies(['client-infos']);

  if(isDemo) return {
    nickname: "Demo",
    role: "2" 
  };

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
    return {
      nickname: "",
      role: "2"
    };
  }
}

export {
  useClientInfos
};