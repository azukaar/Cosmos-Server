const { DefinePlugin } = require("webpack")
const { merge } = require("webpack-merge")
const { resolve } = require("path")
const webpackCommon = require("./webpack.common.js")

module.exports = merge(webpackCommon, {
    mode: "development",
    devtool: "inline-source-map",
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
