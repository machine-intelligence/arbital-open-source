'use strict';

// arb-index directive displays a set of featured domains
app.directive('arbIndex', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/indexPage.html'),
		controller: function($scope) {
			$scope.arb = arb;
			$scope.readTab = 0;
			$scope.writeTab = 0;
			$scope.showJoinSlack = false;
			$scope.slackInvite = {email: ''};

			$scope.joinSlack = function() {
				$scope.showJoinSlack = true;
				$scope.slackInvite.email = arb.userService.user.email;
			};

			$scope.joinSlackSubmit = function() {
				arb.stateService.postDataWithoutProcessing('/json/sendSlackInvite/', $scope.slackInvite, function() {
					arb.userService.user.isSlackMember = true;
				});
				arb.userService.user.isSlackMember = true;
			};

			$scope.selectReadTab = function(tab) {
				$scope.readTab = tab;
			};

			$scope.selectWriteTab = function(tab) {
				$scope.writeTab = tab;
			};
		},
	};
});
