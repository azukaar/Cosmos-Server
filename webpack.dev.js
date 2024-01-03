const { merge } = require("webpack-merge")
const { resolve } = require("path")
const webpackCommon = require("./webpack.common.js")
const webpackProd = require("./webpack.prod.js")

module.exports = merge(process.env.IS_PRODUCTION ? webpackProd : webpackCommon, {
    mode: "development",
    devtool: process.env.useProduction ? undefined : "inline-source-map",
    target: "web",
    devServer: {
        port: 3000,
        hot: true,
        compress: true,
        historyApiFallback: true,
        static: resolve(__dirname, "static"),
        proxy: {
            "/cosmos/api": {
                target: "http://localhost",
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
    }
})
