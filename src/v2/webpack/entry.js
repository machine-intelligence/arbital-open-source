// Our style sheets
require('scss/arbital.scss');

// All AngularJS templates.
var templates = require.context(
		'html',
		true,
		/\.html$/
)
templates.keys().forEach(function(key) {
	templates(key);
});

//require('js/angular.ts');
