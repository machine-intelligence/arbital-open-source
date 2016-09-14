'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// arb-index directive displays the main page
app.directive('arbIndex', function($http, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/indexPage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isTinyScreen = !$mdMedia('gt-xs');

			// Slack stuff
			$scope.showJoinSlackInput = false;
			$scope.showJoinSlackButton = arb.userService.user && !arb.userService.user.isSlackMember;
			if (Cookies.getJSON('isSlackMember')) {
				$scope.showJoinSlackButton = false;
			}

			$scope.slackInvite = {email: ''};

			$scope.joinSlack = function() {
				$scope.showJoinSlackInput = true;
				$scope.slackInvite.email = arb.userService.user.email;
			};

			$scope.joinSlackSubmit = function() {
				arb.stateService.postDataWithoutProcessing('/json/sendSlackInvite/', $scope.slackInvite, function() {
					arb.userService.user.isSlackMember = true;
					Cookies.set('isSlackMember', true);
				});
				arb.userService.user.isSlackMember = true;
			};

			// Tab stuff
			$scope.readTab = 0;
			$scope.writeTab = 0;

			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};

			$scope.selectWriteTab = function(tab) {
				$scope.writeTab = tab;
			};
		},
	};
});
