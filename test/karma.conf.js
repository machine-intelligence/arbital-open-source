// Karma configuration

module.exports = function(config) {
	config.set({

		// base path that will be used to resolve all patterns (eg. files, exclude)
		basePath: '../src/site',

		// frameworks to use
		// available frameworks: https://npmjs.org/browse/keyword/karma-adapter
		frameworks: ['jasmine'],

		plugins: [
			// Karma will require() these plugins
			'karma-jasmine',
			'karma-chrome-launcher',
			'karma-firefox-launcher',
			'karma-safari-launcher',
			'karma-ie-launcher',
			'karma-phantomjs-launcher',
			'karma-ng-html2js-preprocessor'
		],

		// preprocess matching files before serving them to the browser
		// available preprocessors: https://npmjs.org/browse/keyword/karma-preprocessor
		preprocessors: {
			'../../src/**/*.html': ['ng-html2js']
		},

		// list of files / patterns to load in the browser
		files: [
			'../../bower_components/jquery/dist/jquery.js',

			'../../bower_components/angular/angular.js',
			'../../bower_components/angular-mocks/angular-mocks.js',
			'../../bower_components/angular-route/angular-route.js',

			'../../bower_components/MathJax/MathJax.js',

			'../../bower_components/angular-animate/angular-animate.js',
			'../../bower_components/angular-aria/angular-aria.js',
			'../../bower_components/angular-material/angular-material.js',
			'../../bower_components/angular-messages/angular-messages.js',
			'../../bower_components/angular-recursion/angular-recursion.js',
			'../../bower_components/angular-resource/angular-resource.js',
			'../../bower_components/angular-sanitize/angular-sanitize.js',

			'../../src/site/static/**/*.js',
			'../../src/**/*.html',
			'../../test/unit/**/*.js'
		],

		// list of files to exclude
		exclude: [
		],

		// test results reporter to use
		// possible values: 'dots', 'progress'
		// available reporters: https://npmjs.org/browse/keyword/karma-reporter
		reporters: ['progress'],

		// web server port
		port: 9876,

		// enable / disable colors in the output (reporters and logs)
		colors: true,

		// level of logging
		// possible values: config.LOG_DISABLE || config.LOG_ERROR || config.LOG_WARN || config.LOG_INFO || config.LOG_DEBUG
		logLevel: config.LOG_INFO,

		// enable / disable watching file and executing tests whenever any file changes
		autoWatch: true,

		// start these browsers
		// available browser launchers: https://npmjs.org/browse/keyword/karma-launcher
		//browsers: ['Chrome', 'Firefox', 'Safari', 'IE'],
		browsers: ['Chrome', 'Firefox'],
		//browsers: ['PhantomJS'],

		// Continuous Integration mode
		// if true, Karma captures browsers, runs the tests and exits
		singleRun: false,

		// Concurrency level
		// how many browser should be started simultaneous
		concurrency: Infinity
	})
}
