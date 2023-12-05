const { dependencies, peerDependencies } = require("./package.json")
const { IgnorePlugin } = require("webpack")
const { merge } = require("webpack-merge")
const ImageMinimizerPlugin = require("image-minimizer-webpack-plugin")
const CssMinimizerPlugin = require("css-minimizer-webpack-plugin")
const TerserPlugin = require("terser-webpack-plugin")
const webpackCommon = require("./webpack.common.js")

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

module.exports = merge(webpackCommon, {
    mode: "production",
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
                    sourceMap: false,
                    format: {
                        comments: false
                    }
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
})

function listNodeModulesRegExp(deps) {
    return new RegExp(`[\\/]node_modules[\\/]${deps.join("|")}`);
};
