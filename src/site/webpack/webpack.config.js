var path = require('path');
var webpack = require('webpack');

module.exports = {
	plugins: [
	],
	entry: './entry.js',
	output: {
		path: '../static/js',
		// Keep port in sync with pageHandler.go
		publicPath: 'http://localhost:8014/static/js/',
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
