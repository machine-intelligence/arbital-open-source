var base = require('./base.config.js');

module.exports = base.merge({
	devtool: 'source-map',
	module: {
		preLoaders: [
			// Re-process sourcemaps from TypeScript.
			{ test: /\.js$/, loader: "source-map-loader" },
		],
	},
})
