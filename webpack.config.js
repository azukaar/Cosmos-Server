const { dependencies, peerDependencies } = require("./package.json")
const { IgnorePlugin } = require("webpack")
const { BundleAnalyzerPlugin } = require("webpack-bundle-analyzer")
const { DuplicatesPlugin } = require("inspectpack/plugin");
const ImageMinimizerPlugin = require("image-minimizer-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const TerserPlugin = require("terser-webpack-plugin");
const path = require("path")

const dependenciesList = Object.keys(dependencies).concat(Object.keys(peerDependencies))
const cacheGroups = {
    react: listNodeModulesRegExp(
        dependenciesList.filter(dependency => dependency.includes("react"))
            .filter(dependency => ![
                "react-syntax-highlighter",
                "react-element-to-jsx-string",
                "react-intersection-observer",
                "react-device-detect",
                "react-apexcharts",
                "react-copy-to-clipboard"
            ].includes(dependency))
    ),
    highlighter: listNodeModulesRegExp(["react-syntax-highlighter", "react-element-to-jsx-string", "react-copy-to-clipboard"]),
    plot: listNodeModulesRegExp(["react-intersection-observer"]),
    simpleBar: listNodeModulesRegExp(["react-device-detect"]),
    charts: listNodeModulesRegExp(["react-apexcharts", "apexcharts"]),
    mui: listNodeModulesRegExp(dependenciesList.filter(dependency => dependency.includes("mui")))
};

module.exports = {
    entry: "./client/src/index",
    performance: {
        hints: false
    },
    module: {
        rules: [
            {
                test: /\.(ts|ts|js|js|mjs|cjs)x?$/i,
                use: {
                    loader: "babel-loader",
                    options: {
                        cacheCompression: true,
                    }
                },
                exclude: /node_modules/,
            },
            {
                test: /\.s[ac]ss$/i,
                use: [
                    "sass-loader",
                    MiniCssExtractPlugin.loader,
                    {
                        loader: "css-loader", options: {
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
                        loader: "css-loader", options: {
                            importLoaders: 1,
                            modules: true
                        }
                    },
                ],
            },
            {
                test: /\.(jpe?g|png|gif|svg)$/i,
                type: "asset",
            }
        ],
    },
    resolve: {
        extensions: [".*", ".tsx", ".ts", ".jsx", ".js", ".mjs", ".cjs"],
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
    },
    plugins: [
        new MiniCssExtractPlugin(),
        new IgnorePlugin({
            resourceRegExp: /[\d\D]*.demo[\d\D]*/
        }),
        new HtmlWebpackPlugin({
            template: path.join(__dirname, "client", "index.html"),
            filename: "index.[contenthash].html"
        }),
        new BundleAnalyzerPlugin(),
        new DuplicatesPlugin({
            emitErrors: true,
            verbose: true
        })
    ],
    output: {
        path: path.resolve(__dirname, "client/dist"),
        clean: true
    },
    optimization: {
        chunkIds: 'total-size',
        moduleIds: 'size',
        mergeDuplicateChunks: true,
        portableRecords: true,
        sideEffects: true,
        concatenateModules: true,
        splitChunks: {
            chunks: "all",
            cacheGroups
        },
        usedExports: true,
        minimize: true,
        minimizer: [
            new CssMinimizerPlugin(),
            new TerserPlugin({
                minify: TerserPlugin.swcMinify,
                terserOptions: {
                    mangle: true,
                    compress: true,
                    format: {
                        comments: false
                    },
                    sourceMap: false
                }
            }),
            new ImageMinimizerPlugin({
                minimizer: {
                    implementation: ImageMinimizerPlugin.imageminMinify,
                    options: {
                        plugins: [
                            ["gifsicle", { interlaced: true }],
                            ["jpegtran", { progressive: true }],
                            ["optipng", { optimizationLevel: 5 }],
                            [
                                "svgo",
                                {
                                    multipass: true,
                                    plugins: [
                                        "preset-default"
                                    ],
                                },
                            ],
                        ]
                    }
                }
            })
        ]
    }
};

function listNodeModulesRegExp(deps) {
    return new RegExp(`[\\/]node_modules[\\/]${deps.join("|")}`);
};
