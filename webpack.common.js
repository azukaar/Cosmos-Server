const { DuplicatesPlugin } = require("inspectpack/plugin")
const { DefinePlugin } = require("webpack")
const { join } = require("path")
const StatoscopeWebpackPlugin = require('@statoscope/webpack-plugin').default
const MiniCssExtractPlugin = require("mini-css-extract-plugin")
const HtmlWebpackPlugin = require("html-webpack-plugin")

const IS_DEMO = !!process.env.IS_DEMO
const WITH_REPORT = !!process.env.WITH_REPORT
const ANALYZE_DEPS = !!process.env.ANALYZE_DEPS
const ASSET_PATH = process.env.ASSET_PATH || '/'

module.exports = {
    entry: join(__dirname, "client/src/index"),
    output: {
        publicPath: ASSET_PATH,
        path: join(__dirname, "static"),
        clean: true
    },
    plugins: [
        new MiniCssExtractPlugin(),
        new HtmlWebpackPlugin({
            baseUrl: ASSET_PATH,
            template: "client/index.html",
            inject: true,
            minify: true
        }),
        new DefinePlugin({
            "process.env.MODE": JSON.stringify(IS_DEMO ? "demo" : "production")
        })
    ].concat(WITH_REPORT ? [new StatoscopeWebpackPlugin()] : [])
        .concat(ANALYZE_DEPS ? [new DuplicatesPlugin({ emitErrors: true, verbose: true })] : []),
    module: {
        rules: [
            {
                test: /\.(ts|js|mjs|cjs)x?$/i,
                use: ["babel-loader"],
                exclude: /node_modules/,
            },
            {
                test: /\.s[ac]ss$/i,
                use: [
                    "sass-loader",
                    MiniCssExtractPlugin.loader,
                    {
                        loader: "css-loader",
                        options: {
                            importLoaders: 1,
                            modules: true
                        }
                    }
                ],
            },
            {
                test: /\.css$/i,
                use: [
                    MiniCssExtractPlugin.loader,
                    {
                        loader: "css-loader",
                        options: {
                            importLoaders: 1,
                            modules: true
                        }
                    },
                ],
            },
            {
                test: /\.(jpe?g|png|gif|svg)$/i,
                type: "asset/resource",
            }
        ],
    },
    resolve: {
        extensions: [".*", ".tsx", ".ts", ".mts", ".jsx", ".js", ".mjs"],
        fallback: {
            "stream": require.resolve("stream-browserify"),
            "crypto": require.resolve("crypto-browserify"),
            "path": require.resolve("path-browserify"),
            "buffer": require.resolve("buffer/"),
            "util": require.resolve("util/"),
            "fs": false
        },
        alias: {
            "bn.js": require.resolve("bn.js"),
            "isarray": require.resolve("isarray"),
            "level-fix-range": require.resolve("level-fix-range"),
            "object-keys": require.resolve("object-keys"),
            "prr": require.resolve("prr"),
            "react-is": require.resolve("react-is"),
            "safe-buffer": require.resolve("safe-buffer"),
            "string_decoder": require.resolve("string_decoder"),
            "xtend": require.resolve("xtend"),
            "framer-motion": require.resolve("framer-motion")
        }
    }
}