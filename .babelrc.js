module.exports = {
    compact: false,
    presets: [
        "@babel/env",
        "@babel/react",
        "@babel/preset-typescript"
    ],
    plugins: [
        "@babel/plugin-transform-runtime",
        ["import", {
            libraryName: "redux",
            libraryDirectory: "src",
            camel2DashComponentName: false,
            transformToDefaultImport: true
        }, "redux"]
    ]
}
