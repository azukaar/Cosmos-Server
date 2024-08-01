// has number
const hasNumber = (number) => new RegExp(/[0-9]/).test(number);

// has mix of small and capitals
const hasMixed = (number) => new RegExp(/[a-z]/).test(number) && new RegExp(/[A-Z]/).test(number);

// has special chars
const hasSpecial = (number) => new RegExp(/[~!@#$%\^&\*\(\)_\+=\-\{\[\}\]:;"'<,>\?\/]/).test(number);

// set color based on password strength
export const strengthColor = (count, t) => { 
    if (count < 2) return { label: t('auth.genPwdStrength.poor'), color: 'error.main' };
    if (count < 3) return { label: t('auth.genPwdStrength.weak'), color: 'warning.main' };
    if (count < 4) return { label: t('auth.genPwdStrength.normal'), color: 'warning.dark' };
    if (count < 5) return { label: t('auth.genPwdStrength.good'), color: 'success.main' };
    if (count < 6) return { label: t('auth.genPwdStrength.strong'), color: 'success.dark' };
    return { label: t('auth.genPwdStrength.poor'), color: 'error.main' };
};

// password strength indicator
export const strengthIndicator = (number) => {
    let strengths = 0;
    if (number.length > 9) strengths += 1;
    if (number.length > 11) strengths += 1;
    if (hasNumber(number)) strengths += 1;
    if (hasSpecial(number)) strengths += 1;
    if (hasMixed(number)) strengths += 1;
    return strengths;
};
