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

			$scope.fakeProjectPages = [
				{isRedLink: true, alias: 'Polynomial'},
				{isRedLink: true, alias: 'Isomorphism'},
				{isRedLink: true, alias: 'Finite sets'},
				{isRedLink: true, alias: 'Disjoint union of finite sets'},
				{title: 'The empty set can be described entirely by its universal property', quality: 'Stub'},
				{title: 'Least upper bound and greatest lower bounds', quality: 'Stub'},
				{title: 'The universal properties of the LUB and GLB', quality: 'Stub'},
			];
		},
	};
});
