var webpack = require('webpack');
var base = require('./base.config.js');

module.exports = Object.assign(base, {
	plugins: [
		new webpack.optimize.UglifyJsPlugin({
			mangle: false, // Mangling currently breaks Angular.
			compress: {warnings: false} // The warnings are too verbose.
		}),
	]
})
