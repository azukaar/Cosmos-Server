import { useEffect, useState } from 'react';
import * as API from '../api';
import proFeatures from '../pro';

// Whether this bundle ships the Pro components; community builds stub pro/index.js with nulls.
export const isProBuild = () => !!(proFeatures.isPro && proFeatures.isPro());

// The licence answer is resolved once per session and shared. null = not resolved yet.
let licenceState = null;
let licencePromise = null;
const listeners = new Set();

function resolveLicence() {
  if (licencePromise) return licencePromise;

  licencePromise = API.getStatus().then((res) => {
    licenceState = !!(res && res.data && res.data.Licence);
    listeners.forEach((notify) => notify(licenceState));
    return licenceState;
  }).catch(() => {
    // A failing status call is not a licence: treat it as "free".
    licenceState = false;
    listeners.forEach((notify) => notify(licenceState));
    return false;
  });

  return licencePromise;
}

export function useHasLicence() {
  const [licence, setLicence] = useState(licenceState);

  useEffect(() => {
    if (licenceState !== null) {
      setLicence(licenceState);
      return;
    }
    listeners.add(setLicence);
    resolveLicence();
    return () => {
      listeners.delete(setLicence);
    };
  }, []);

  return licence;
}

// True only with the Pro bundle present AND a server licence; null while the licence is still unknown.
export function useIsPro() {
  const licence = useHasLicence();

  if (!isProBuild()) return false;
  return licence === null ? null : licence;
}
