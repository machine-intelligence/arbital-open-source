var path = require('path');
var webpack = require('webpack');

module.exports = {
	plugins: [
		new webpack.optimize.UglifyJsPlugin({
			mangle: false, // Mangling currently breaks Angular.
			compress: {warnings: false} // The warnings are too verbose.
		}),
	],
	entry: './entry.js',
	output: {
		path: '../static/js',
		filename: 'bundle.js',
	},
	module: {
		noParse: /js\/lib\//,
		loaders: [
			// TypeScript
			{ test: /\.tsx?$/, loader: 'ts-loader' },

			// SCSS
			{ test: /\.scss$/, loaders: ['style', 'css', 'sass'] }
		]
	},
	resolve: {
		root: [
			path.resolve('../static'),
			path.resolve('../../node_modules'),
		],
	},
};
