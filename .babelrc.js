module.exports = {
    compact: false,
    presets: [
        "@babel/env",
        "@babel/preset-typescript",
        ["@babel/react", { "runtime": "automatic" }]
    ],
    plugins: [
        ["@babel/plugin-transform-runtime"],
        ["import", {
            libraryName: "redux",
            libraryDirectory: "src",
            camel2DashComponentName: false,
            transformToDefaultImport: true
        }, "redux"]
    ]
}
