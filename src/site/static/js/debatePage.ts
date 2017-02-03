'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// arb-debate directive displays the project page
app.directive('arbDebate', function($http, $mdDialog, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/debatePage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.pageId = '7qq';
			$scope.page = arb.stateService.pageMap[$scope.pageId];
		},
		link: function(scope: any, element, attrs) {
			console.log('in link');

			// Create a dialog for (resuming) editing a new claim
			// ROGTODO: figure out whether to use resumeClaimPageId or not? (and how)
			var resumeClaimPageId = undefined;
			scope.showNewClaimDialog = function(event) {
				console.log('in showNewClaimDialog');

				var title = undefined;
				$mdDialog.show({
					templateUrl: versionUrl('static/html/editClaimDialog.html'),
					controller: 'EditClaimDialogController',
					autoWrap: false,
					targetEvent: event,
					locals: {
						resumePageId: resumeClaimPageId,
						title: title,
						originalPage: scope.page,
					},
				});
				return false;
			};
		},
	};
});
