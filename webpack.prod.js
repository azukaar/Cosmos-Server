const { dependencies, peerDependencies } = require("./package.json")
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
                "react-copy-to-clipboard",
                "react-simplebar",
                "react-material-ui-carousel"
            ].includes(dependency))
    ),
    highlighter: listNodeModulesRegExp(["react-syntax-highlighter", "react-element-to-jsx-string", "react-copy-to-clipboard"]),
    plot: listNodeModulesRegExp(["react-intersection-observer"]),
    simpleBar: listNodeModulesRegExp(["react-device-detect", "react-simplebar", "simplebar"]),
    charts: listNodeModulesRegExp(["react-apexcharts", "apexcharts"]),
    mui: listNodeModulesRegExp(dependenciesList
        .filter(dependency => dependency.includes("mui"))
        .concat(["@mui/system", "react-material-ui-carousel"])
    ),
    yaml: listNodeModulesRegExp(["js-yaml"]),
    crypto: listNodeModulesRegExp(["crypto-browserify"]),
    coreJS: listNodeModulesRegExp(["core-js"])
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
                extractComments: {
                    condition: /^\**!|@preserve|@license|@cc_on/i,
                    filename: fileData => {
                        return `${fileData.filename}.LICENSE.txt${fileData.query}`;
                    },
                    banner: licenseFile => {
                        return `License information can be found in ${licenseFile}`;
                    },
                },
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
