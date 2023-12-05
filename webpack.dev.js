const { DefinePlugin } = require("webpack")
const { merge } = require("webpack-merge")
const { resolve } = require("path")
const webpackCommon = require("./webpack.common.js")
const webpackProd = require("./webpack.prod.js")

module.exports = merge(process.env.useProduction ? webpackProd : webpackCommon, {
    mode: "development",
    devtool: !process.env.useProduction ? "inline-source-map" : undefined,
    target: "web",
    devServer: {
        port: 3000,
        hot: true,
        compress: true,
        historyApiFallback: true,
        static: resolve(__dirname, "static"),
        proxy: {
            "/cosmos/api": {
                target: `http://${process.env.PROXY_HOST || "localhost"}:${process.env.PROXY_PORT}`,
                logLevel: "debug",
                secure: false,
                ws: true
            }
        },
        client: {
            overlay: {
                errors: true,
                warnings: true
            }
        }
    },
    plugins: [
        new DefinePlugin({
            "import.meta.env.MODE": JSON.stringify("demo")
        })
    ]
})
