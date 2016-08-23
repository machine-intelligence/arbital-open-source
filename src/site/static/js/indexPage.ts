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

			arb.stateService.postData('/json/project/', {},
				function(response) {
					// Compute rows to display all the pages in the project
					var aliasRows = response.result.projectData.aliasRows.map(function(aliasRow) {
						return {isRedLink: true, alias: aliasRow.alias};
					});
					var pageRows = response.result.projectData.pageIds.map(function(pageId) {
						var page = arb.stateService.getPage(pageId);
						page.qualityTag = arb.pageService.getQualityTagId(page.tagIds);

						if (page.qualityTag == 'unassessed') {
							page.qualityTag = 'Unassessed';
						}

						return page;
					});
					$scope.projectPageRows = aliasRows.concat(pageRows);

					// Compute recent changes rows
					$scope.changeLogModeRows = [];
					let acceptedChangeLogTypes = {newEditProposal: true, newEdit: true, deletePage: true, revertEdit: true};
					for (let n = 0; n < response.result.projectData.pageIds.length; n++) {
						let page = arb.stateService.pageMap[response.result.projectData.pageIds[n]];
						for (let i = 0; i < page.changeLogs.length; i++) {
							let changeLog = page.changeLogs[i];
							if (!acceptedChangeLogTypes[changeLog.type]) continue;
							$scope.changeLogModeRows.push({
								rowType: changeLog.type,
								activityDate: changeLog.createdAt,
								changeLog: changeLog,
							});
						}
					}

					// Compute "X changes by Y authors in last week" text
					let changeCountLastWeek = 0;
					let authorIdsSet = {};
					let now = moment.utc();
					for (let n = 0; n < $scope.changeLogModeRows.length; n++) {
						let changeLog = $scope.changeLogModeRows[n].changeLog;
						if (now.diff(moment.utc(changeLog.createdAt), 'days') > 7) {
							break;
						}
						authorIdsSet[changeLog.userId] = true;
						changeCountLastWeek++;
					}
					let authorCountLastWeek = Object.keys(authorIdsSet).length;
					$scope.changesCountText = '' + changeCountLastWeek;
					if ($scope.changesCountLastWeek == 1) {
						$scope.changesCountText += ' change';
					} else {
						$scope.changesCountText += ' changes';
					}
					$scope.changesCountText += ' by ' + authorCountLastWeek;
					if (authorCountLastWeek == 1) {
						$scope.changesCountText += ' author';
					} else {
						$scope.changesCountText += ' authors';
					}
					$scope.changesCountText += ' last week';
				});
		},
	};
});
