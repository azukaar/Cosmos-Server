const { IgnorePlugin } = require("webpack")
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
const TerserPlugin = require("terser-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const path = require('path');

module.exports = {
    entry: './client/src/index',
    module: {
        rules: [
            {
                test: /\.(ts|tsx)$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.(js|jsx)$/,
                use: "babel-loader",
                exclude: /node_modules/,
            },
            {
                test: /\.s[ac]ss$/i,
                use: [
                    MiniCssExtractPlugin.loader,
                    { loader: "css-loader", options: { modules: true } },
                    "sass-loader",
                ],
            },
            {
                test: /\.css$/i,
                use: ["style-loader", { loader: "css-loader", options: { modules: true } },],
            },
            {
                test: /\.(jpe?g|png|gif|svg)$/i,
                loader: 'file-loader',
            }
        ],
    },
    resolve: {
        extensions: ['.*', '.tsx', '.ts', '.jsx', '.js'],
        fallback: {
            "stream": require.resolve("stream-browserify"),
            "crypto": require.resolve("crypto-browserify"),
            "path": require.resolve("path-browserify"),
            "fs": require.resolve("browserify-fs"),
            "buffer": require.resolve("buffer/"),
            "util": require.resolve("util/"),
        }
    },
    plugins: [
        new IgnorePlugin({
            resourceRegExp: /[\d\D]*.demo[\d\D]*/
        }),
        new HtmlWebpackPlugin({
            template: path.join(__dirname, "client", "index.html"),
        }),
        new BundleAnalyzerPlugin(),
    ],
    output: {
        path: path.resolve(__dirname, 'client/dist'),
    },
    optimization: {
        chunkIds: "total-size",
        moduleIds: "size",
        mergeDuplicateChunks: true,
        portableRecords: true,
        sideEffects: true,
        minimizer: [
            new CssMinimizerPlugin(),
            new TerserPlugin({
                minify: TerserPlugin.swcMinify,
                parallel: true,
                extractComments: "all",
                terserOptions: {
                    mangle: true,
                    compress: true,
                    sourceMap: false,
                }
            })
        ]
    }
};
