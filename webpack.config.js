const { IgnorePlugin } = require('webpack')
const { BundleAnalyzerPlugin } = require('webpack-bundle-analyzer')
const { DuplicatesPlugin } = require("inspectpack/plugin");
const TerserPlugin = require('terser-webpack-plugin');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const CssMinimizerPlugin = require('css-minimizer-webpack-plugin');
const ESLintPlugin = require('eslint-webpack-plugin');

const path = require('path');

module.exports = {
    entry: './client/src/index',
    module: {
        rules: [
            {
                test: /\.(ts|tsx|js|jsx|mjs|cjs)$/i,
                use: {
                    loader: 'babel-loader',
                    options: {
                        cacheCompression: true
                    }
                },
                exclude: /node_modules/,
            },
            {
                test: /\.s[ac]ss$/i,
                use: [
                    MiniCssExtractPlugin.loader,
                    { loader: 'css-loader', options: { modules: true } },
                    'sass-loader',
                ],
            },
            {
                test: /\.css$/i,
                use: ['style-loader', { loader: 'css-loader', options: { modules: true } },],
            },
            {
                test: /\.(jpe?g|png|gif|svg)$/i,
                loader: 'file-loader',
            }
        ],
    },
    resolve: {
        extensions: ['.*', '.tsx', '.ts', '.jsx', '.js', '.mjs', '.cjs'],
        fallback: {
            'stream': require.resolve('stream-browserify'),
            'crypto': require.resolve('crypto-browserify'),
            'path': require.resolve('path-browserify'),
            'fs': require.resolve('browserify-fs'),
            'buffer': require.resolve('buffer/'),
            'util': require.resolve('util/'),
        },
        alias: {
            'bn.js': require.resolve('bn.js'),
            'isarray': require.resolve('isarray'),
            'level-fix-range': require.resolve('level-fix-range'),
            'object-keys': require.resolve('object-keys'),
            'prr': require.resolve('prr'),
            'react-is': require.resolve('react-is'),
            'safe-buffer': require.resolve('safe-buffer'),
            'string_decoder': require.resolve('string_decoder'),
            'xtend': require.resolve('xtend')
        }
    },
    plugins: [
        new ESLintPlugin(),
        new IgnorePlugin({
            resourceRegExp: /[\d\D]*.demo[\d\D]*/
        }),
        new HtmlWebpackPlugin({
            template: path.join(__dirname, 'client', 'index.html'),
            filename: 'index.[contenthash].html'
        }),
        new BundleAnalyzerPlugin(),
        new DuplicatesPlugin({
            emitErrors: false,
            verbose: true
        })
    ],
    output: {
        path: path.resolve(__dirname, 'client/dist'),
    },
    optimization: {
        chunkIds: 'total-size',
        moduleIds: 'size',
        mergeDuplicateChunks: true,
        portableRecords: true,
        sideEffects: true,
        splitChunks: {
            chunks: 'all',
        },
        minimizer: [
            new CssMinimizerPlugin(),
            new TerserPlugin({
                minify: TerserPlugin.swcMinify,
                terserOptions: {
                    mangle: true,
                    compress: true,
                    sourceMap: false,
                }
            }),
            new TerserPlugin({
                minify: TerserPlugin.uglifyJsMinify
            }),
            new TerserPlugin({
                minify: TerserPlugin.esbuildMinify
            }),
            new TerserPlugin({
                minify: TerserPlugin.terserMinify
            }),

        ]
    }
};
