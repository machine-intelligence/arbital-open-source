'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// arb-debate directive displays the project page
app.directive('arbDebate', function($http, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/debatePage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isTinyScreen = !$mdMedia('gt-xs');
		},
	};
});
