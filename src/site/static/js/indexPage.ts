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

			// Code snippet for showing project logs
			/*arb.stateService.postData('/json/project/', {}, function(data) {
				console.log(data);
				console.log(data.result.projectData.pageIds);
				$scope.changeLogModeRows = [];
				for (let n = 0; n < data.result.projectData.pageIds.length; n++) {
					let page = arb.stateService.pageMap[data.result.projectData.pageIds[n]];
					for (let i = 0; i < page.changeLogs.length; i++) {
						$scope.changeLogModeRows.push({
							rowType: page.changeLogs[i].type,
							activityDate: page.changeLogs[i].createdAt,
							changeLog: page.changeLogs[i],
						});
					}
				}
				//$scope.changeLogModeRows.sort(function(a,b) {
					//return b.createdAt.localeCompare(a.createdAt);
				//});
				console.log($scope.changeLogModeRows);

				// Compute "X changes by Y authors in last week"
				$scope.changeCountLastWeek = 0;
				let authorIdsSet = {};
				let now = moment.utc();
				for (let n = 0; n < $scope.changeLogModeRows.length; n++) {
					let changeLog = $scope.changeLogModeRows[n].changeLog;
					if (now.diff(moment.utc(changeLog.createdAt), 'days') > 7) {
						break;
					}
					authorIdsSet[changeLog.userId] = true;
					$scope.changeCountLastWeek++;
				}
				$scope.authorCountLastWeek = Object.keys(authorIdsSet).length;
				console.log($scope.changeCountLastWeek);
				console.log($scope.authorCountLastWeek);
			});*/
		},
	};
});
