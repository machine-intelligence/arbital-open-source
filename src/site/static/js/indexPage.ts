'use strict';

import app from './angular.ts';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/indexPage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.isTinyScreen = !$mdMedia('gt-xs');

			$scope.readTab = 0;
			$scope.writeTab = 0;
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

			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};

			$scope.selectWriteTab = function(tab) {
				$scope.writeTab = tab;
			};

			$scope.stubs = [
				'We only care about things up to isomorphism',
				'We can describe objects based entirely on how they interact with other objects',
				'The category of finite sets',
				'The empty set',
				'Disjoint union of finite sets',
				'Union of finite sets',
				'The empty set can be described enitrely by its universal property',
				'The union and product can be described entirely by their universal properties, up to isomorphism',
				'Least upper bound and greatest lower bounds',
				'The universal properties of the LUB and GLB',
				'This kind of property crops up all over the place'
			];
		},
	};
});
