import { Button } from "@mui/material";
import { useTheme } from '@mui/material/styles';

export const randomString = (length) => {
  let text = "";
  const possible =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  for (let i = 0; i < length; i++)
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  return text;
}

export function isDomain(hostname) {
  // Regular expression to check if it's an IP address
  const ipPattern = /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/;

  // Remove port if there is one
  hostname = hostname.replace(/:\d+$/, '');

  // Check if the hostname is an IP address
  if (ipPattern.test(hostname)) {
      return false;
  }

  // Check if the hostname is "localhost"
  if (hostname === 'localhost') {
      return false;
  }

  return true;
}

export const debounce = (func, wait) => {
    let timeout;
    return function (...args) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), wait);
    };
  };

export const redirectTo = (url) => {
  window.location.href = url;
}

export const redirectToLocal = (url) => {
  let redirectUrl = new URL(url, window.location.href);
  let currentLocation = window.location;
  if (redirectUrl.origin != currentLocation.origin){
    throw new Error("URL must be local");
  }
  window.location.href = url;
}

// Validates a crontab expression. Accepts:
//   - 5-field standard:  minute hour dom month dow
//   - 6-field seconds:   second minute hour dom month dow
//   - descriptors:       @daily, @hourly, @yearly, @weekly, @monthly, @every 10m
//   - timezone prefix:   CRON_TZ=Europe/Paris <expr>
// Returns true when the field count and values are valid cron syntax.
export const isValidCrontab = (crontab) => {
  if (!crontab) return false;
  let expr = crontab.trim();
  if (expr.startsWith('@')) return true;

  // strip a leading TZ= / CRON_TZ= prefix
  const m = expr.match(/^(TZ|CRON_TZ)=[^\s]+\s+(.*)$/);
  if (m) expr = m[2].trim();

  const parts = expr.split(/\s+/).filter(Boolean);
  if (parts.length !== 5 && parts.length !== 6) return false;

  const ranges = {
    sec: [0, 59],
    min: [0, 59],
    hour: [0, 23],
    dom: [1, 31],
    month: [1, 12],
    dow: [0, 7], // 0 and 7 both mean Sunday
  };
  const names = parts.length === 6
    ? ['sec','min','hour','dom','month','dow']
    : ['min','hour','dom','month','dow'];

  const parseRange = (field) => {
    if (field === '*') return true;
    // steps: */5 or 1-10/2 or 5/15
    const base = field.split('/')[0];
    if (base === '*') return true;
    if (base.includes(',')) return base.split(',').every(parseRange);
    if (base.includes('-')) {
      const [a, b] = base.split('-');
      return !isNaN(a) && !isNaN(b) && parseInt(a) >= 0 && parseInt(b) >= 0 && parseInt(a) <= parseInt(b);
    }
    return !isNaN(base);
  };

  return parts.every((p, i) => {
    if (!parseRange(p)) return false;
    // validate literal values are within range (ignore wildcards/steps/ranges)
    const range = ranges[names[i]];
    const vals = p.split('/')[0];
    if (vals.includes(',')) {
      return vals.split(',').every(v => {
        if (v.includes('-')) return true; // range checked above
        const n = parseInt(v);
        return !isNaN(n) && n >= range[0] && n <= range[1];
      });
    }
    if (vals.includes('-') || vals === '*') return true;
    const n = parseInt(vals);
    return !isNaN(n) && n >= range[0] && n <= range[1];
  });
};

export const crontabToText = (crontab, t) => {
  if (!crontab) return t('mgmt.cron.invalidCron');

  let expr = crontab.trim();
  let tzPrefix = '';
  const m = expr.match(/^(TZ|CRON_TZ)=([^\s]+)\s+(.*)$/);
  if (m) {
    tzPrefix = `${m[1]}=${m[2]} `;
    expr = m[3].trim();
  }

  // Descriptors
  if (expr.startsWith('@')) {
    const map = {
      '@yearly': t('mgmt.cron.descriptor.yearly'),
      '@annually': t('mgmt.cron.descriptor.yearly'),
      '@monthly': t('mgmt.cron.descriptor.monthly'),
      '@weekly': t('mgmt.cron.descriptor.weekly'),
      '@daily': t('mgmt.cron.descriptor.daily'),
      '@midnight': t('mgmt.cron.descriptor.daily'),
      '@hourly': t('mgmt.cron.descriptor.hourly'),
      '@every': t('mgmt.cron.descriptor.every'),
    };
    const key = expr.startsWith('@every ') ? '@every' : expr;
    return tzPrefix + (map[key] ? (key === '@every' ? `${map[key]} ${expr.slice(7)}` : map[key]) : expr);
  }

  const parts = expr.split(/\s+/).filter(Boolean);
  if (parts.length !== 5 && parts.length !== 6) {
    return t('mgmt.cron.invalidCron');
  }

  // Canonical 6-field view: seconds first. 5-field gets an implicit "0" seconds.
  const [second, minute, hour, dayOfMonth, month, dayOfWeek] = parts.length === 6 ? parts : ['0', ...parts];

  const DOW = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
  const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

  const wild = (f) => f === '*';
  const lit = (f) => /^\d+$/.test(String(f).trim());
  const num = (f) => parseInt(f);
  const plural = (n) => (n === 1 ? '' : 's');
  const zero = (f) => wild(f) || (lit(f) && num(f) === 0);
  const step = (f) => { const i = String(f).lastIndexOf('/'); return i >= 0 ? f.slice(i + 1) : null; };
  const pad = (n) => String(parseInt(n)).padStart(2, '0');
  const listOf = (f, fn) => f.includes(',') ? f.split(',').map(fn).join(', ') : fn(f);

  const ordinal = (nb) => {
    nb = parseInt(nb);
    if (nb === 1 || nb === 21 || nb === 31) return `${nb}st`;
    if (nb === 2 || nb === 22) return `${nb}nd`;
    if (nb === 3 || nb === 23) return `${nb}rd`;
    return `${nb}th`;
  };

  const dowName = (v) => DOW[parseInt(v) % 7]; // POSIX: 0/7=Sun, 1=Mon .. 6=Sat
  const dowText = (field) => {
    if (field.includes('-')) {
      const [a, b] = field.split('-').map(Number);
      return `${dowName(a)} to ${dowName(b)}`;
    }
    return listOf(field, dowName);
  };

  // ---------------- Time of day ----------------
  let timeClause = '';
  const hStep = hour.includes('/'), mStep = minute.includes('/'), sStep = second.includes('/');

  if (lit(hour) && lit(minute) && !hStep && !mStep) {
    let c = `${pad(hour)}:${pad(minute)}`;
    if (lit(second) && num(second) !== 0) c += `:${pad(second)}`;
    timeClause = `at ${c}`;
  } else if (wild(hour) && wild(minute) && !mStep) {
    if (zero(second)) timeClause = 'every minute';
    else if (sStep) timeClause = `every ${step(second)} seconds`;
    else timeClause = `every ${second} seconds`;
  } else if (wild(minute) && !mStep) {
    if (hStep) timeClause = `every ${step(hour)} hours`;
    else if (lit(hour)) timeClause = `every hour at ${pad(hour)}:00`;
    else timeClause = 'every hour';
  } else if (hour === '*' && lit(minute)) {
    // Repeats every hour at a fixed minute: "0 * * * *" -> every hour at :00.
    if (minute === '0') timeClause = 'every hour at :00';
    else if (num(minute) === 0) timeClause = 'every hour at :00';
    else timeClause = `every hour at :${pad(minute)}`;
  } else if (hStep) {
    // A repeating hour cadence with a fixed minute: "15 */3 * * *".
    timeClause = (minute === '0' || num(minute) === 0)
      ? `every ${step(hour)} hours`
      : `every ${step(hour)} hours at :${pad(minute)}`;
  } else if (lit(hour) && wild(minute)) {
    timeClause = `every hour at ${pad(hour)}:00`;
  } else if (mStep) {
    timeClause = `every ${step(minute)} minutes`;
  } else {
    const bits = [];
    if (!wild(hour)) bits.push(hStep ? `every ${step(hour)} hours` : `${hour} hour${plural(num(hour))}`);
    if (!wild(minute)) bits.push(mStep ? `every ${step(minute)} minutes` : `${minute} min`);
    if (!zero(second) && !wild(second)) bits.push(`${second} sec`);
    timeClause = bits.length ? `at ${bits.join(' and ')}` : '';
  }

  // ---------------- Calendar ----------------
  let dayClause = '';
  const domW = wild(dayOfMonth), monW = wild(month), dowW = wild(dayOfWeek);
  const monStep = month.includes('/');

  if (domW && monW && dowW) {
    dayClause = ''; // daily
  } else if (monStep) {
    dayClause = `every ${step(month)} months`;
    if (!domW) dayClause += ` on the ${listOf(dayOfMonth, ordinal)}`;
  } else if (dayOfMonth.includes('/')) {
    dayClause = `every ${step(dayOfMonth)} days`;
    if (!monW) dayClause += ` in ${MONTHS[num(month) - 1]}`;
  } else if (!dowW) {
    dayClause = `every week on ${dowText(dayOfWeek)}`;
    if (!monW) dayClause += ` in ${MONTHS[num(month) - 1]}`;
  } else if (!domW && !monW) {
    dayClause = `every ${MONTHS[num(month) - 1]} ${ordinal(dayOfMonth)}`;
  } else if (!domW) {
    dayClause = `every month on the ${listOf(dayOfMonth, ordinal)}`;
  } else if (!monW) {
    dayClause = `every ${MONTHS[num(month) - 1]}`;
  } else {
    dayClause = 'every month';
  }

  // ---------------- Assemble ----------------
  let out;
  if (dayClause === '') {
    // A bare "at hh:mm" means daily.
    out = timeClause.startsWith('at ')
      ? `every day ${timeClause}`
      : timeClause;
  } else {
    out = [dayClause, timeClause].filter(Boolean).join(' ');
  }

  out = (tzPrefix + out).trim();
  return out === '' ? t('mgmt.cron.invalidCron') : out;
}

export const PascalToSnake = (str) => {
  return str.replace(/[\w]([A-Z])/g, function(m) {
    return m[0] + "_" + m[1];
  }).toLowerCase();
}

export const getCurrencyFromLanguage = () => {
  let language = window.navigator.userLanguage || window.navigator.language;
  // language = language.split('-')[0]; // Get language code without region
  
  const currencyMap = {
    en: 'USD', // English (assuming US English as default)
    'en-US': 'USD', // US English
    'en-GB': 'GBP', // British English
    de: 'EUR', // German
    fr: 'EUR', // French
    es: 'EUR', // Spanish
    it: 'EUR', // Italian
    pt: 'EUR', // Portuguese
    nl: 'EUR', // Dutch
  };

  return currencyMap[language] || 'USD'; // Default to USD if no match
};


export const DarkModeSwitch = ({light, dark}) => {
  const theme = useTheme();
  const isLight = theme.palette.mode === 'light';

  return isLight ? light : dark;
}