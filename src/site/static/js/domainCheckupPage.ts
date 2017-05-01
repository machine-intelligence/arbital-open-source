'use strict';

import app from './angular.ts';
import {arraysSortFn} from './util.ts';

// arb-index directive displays the main page
app.directive('arbDomainCheckup', function($http, $mdMedia, $mdToast, arb) {
	return {
		templateUrl: versionUrl('static/html/domainCheckup.html'),
		scope: {
			domainData: '=',
		},
		controller: function($scope) {
			if (!(arb.userService.user.id == '2' || arb.userService.user.id == '76')) {
				arb.urlService.goToUrl('/');
			}
			$scope.arb = arb;
			$scope.rows = [];

			arb.stateService.postData('/json/readMode/', {
					type: 'new',
					numPagesToLoad: 1000,
					domainIdConstraint: $scope.domainData.id,
				},
				function(data) {
					$scope.domainPageIds = data.result.modeRows.map(function(modeRow) {
						return modeRow.pageId;
					});
					$scope.loadParentDataForAllPages();
				});

			$scope.loadParentDataForAllPages = function() {
				$scope.domainPageIds.forEach(function(pageId) {
					arb.pageService.loadEdit({pageAlias: pageId, success: function() {
						$scope.loadDomainDataForParents(pageId);
					}});
				});
			};

			$scope.loadDomainDataForParents = function(pageId) {
				var parents = arb.stateService.pageMap[pageId].parentIds;
				if (parents.length == 0) {
					return;
				}

				var firstParent = parents[0];
				arb.pageService.loadEdit({pageAlias: firstParent, success: function(){
					var page = arb.stateService.pageMap[pageId];
					if (page.parentIds.length == 0) {
						$scope.rows.push({pageId: pageId});
						return;
					}
					var firstParentPage = arb.stateService.pageMap[firstParent];
					var domainOfFirstParent = arb.stateService.domainMap[firstParentPage.editDomainId];
					$scope.rows.push({
						pageId: pageId,
						parentIds: page.parentIds,
						domainOfFirstParent: domainOfFirstParent
					});
				}});
			};

			$scope.updateDomain = function(pageId, newDomainId) {
				var page = arb.stateService.pageMap[pageId];
				page.editDomainId = newDomainId;
				arb.pageService.savePageInfo(page, function(err) {
					if (!err) {
						$mdToast.show($mdToast.simple().textContent('Success'));
						$scope.rows = $scope.rows.filter(function(row) {
							return row.pageId != pageId;
						});
					}
				});
			}
		},
	};
});
