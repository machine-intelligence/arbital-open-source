'use strict';

// arb-achievements-page hosts the arb-achievements-panel
app.directive('arbAchievementsPage', function($http, pageService, userService) {
	return {
		templateUrl: 'static/html/achievements.html',
	};
});
