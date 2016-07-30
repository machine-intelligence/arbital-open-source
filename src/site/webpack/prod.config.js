var webpack = require('webpack');
var base = require('./base.config.js');

module.exports = base.merge({
	plugins: [
		new webpack.optimize.UglifyJsPlugin({
			mangle: false, // Mangling currently breaks Angular.
			compress: {warnings: false} // The warnings are too verbose.
		}),
	]
})
