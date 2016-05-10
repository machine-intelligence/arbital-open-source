'use strict';

// arb-read-mode-page directive displays a list of hot pages, recommended for reading
app.directive('arbReadModePage', function() {
	return {
		templateUrl: 'static/html/readModePage.html',
		scope: {
			hotPageIds: '=',
		},
	};
});
