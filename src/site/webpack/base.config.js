var path = require('path');

module.exports = {
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
			{ test: /\.scss$/, loaders: ['style', 'css', 'sass'] },

			// Angular Templates
			{
				test: /\.html$/,
				loader: 'ngtemplate?relativeTo=site/!html',
			},
		]
	},
	resolve: {
		root: [
			path.resolve('../static'),
			path.resolve('../../node_modules'),
		],
	},
};
