'use strict';

// arb-list-panel is a directive that displays a list
app.directive('arbListPanel', function($http, arb) {
	return {
		templateUrl: 'static/html/listPanel.html',
		scope: {
			numToDisplay: '=',
			isFullPage: '=',
			hideTitle: '=',
			modeRows: '=',
		},
	};
});