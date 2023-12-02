const reactAll = require('eslint-plugin-react/configs/all');
const globals = require('globals');

module.exports = [
    {
        files: ['client/src/**/*.{js,mjs,cjs,jsx,mjsx,ts,tsx,mtsx}'],
        ...reactAll,
        languageOptions: {
            ...reactAll.languageOptions,
            globals: {
                ...globals.serviceworker,
                ...globals.browser,
            },
        },
        rules: {
            'react/jsx-uses-react': 'error',
            'react/jsx-uses-vars': 'error',
            'no-restricted-imports': [
                "error",
                {
                  "patterns": ["@mui/*/*/*"]
                }
            ]
        },
    },
];