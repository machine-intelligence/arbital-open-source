'use strict';

import app from './angular.ts';
import {anyUrlMatch} from './util.ts';
import {aliasMatch} from './markdownService.ts';

// Directive for the Feed page.
app.directive('arbFeedPage', function($timeout, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/feedPage.html'),
		scope: {
			feedRows: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.resetSubmission = function() {
				$scope.isSubmittingLink = false;
				$scope.submission = {
					url: '',
					title: '',
					pageId: '',
				};
			}
			$scope.resetSubmission();

			$scope.showSubmissionForm = function() {
				$scope.isSubmittingLink = true;
				$timeout(function() {$('.submit-link-url-input').focus();});
			};

			// Track page ids we tried fetching from the server.
			let attemptedPageIds = {};
			// Called when the url input changes
			$scope.submissionUrlChanged = function() {
				$scope.externalUrlIsDupe = false;
				$scope.externalUrlOriginalPageID = "";

				if (!$scope.submission.url) {
					return;
				}

				// Figure out if the url is to an Arbital page
				let arbitalUrlRegexp = new RegExp(arb.urlService.getTopLevelDomain() + '/p/' + aliasMatch, 'g');
				let lensUrlRegexp = new RegExp('[&?]l=' + aliasMatch, 'g');
				$scope.submission.pageId = '';
				let matches = arbitalUrlRegexp.exec($scope.submission.url);
				if (!!matches) {
					$scope.submission.pageId = matches[1];
					matches = lensUrlRegexp.exec($scope.submission.url);
					if (matches) {
						$scope.submission.pageId = matches[1];
					}

					// Get the title for the page
					if ($scope.submission.pageId in arb.stateService.pageMap) {
						$scope.submission.title = arb.stateService.pageMap[$scope.submission.pageId].title;
						$scope.submission.pageId = arb.stateService.pageMap[$scope.submission.pageId].pageId;
					} else if (!($scope.submission.pageId in attemptedPageIds)) {
						attemptedPageIds[$scope.submission.pageId] = true;
						arb.pageService.loadTitle($scope.submission.pageId, {
							silentFail: true,
							success: function() {
								$scope.submissionUrlChanged();
							}
						});
					}
				} else {
					let matches = anyUrlMatch.exec($scope.submission.url);
					if (!!matches) {
						handleNewExternalUrl(matches[0]);
					}
				}
			};

			let handleNewExternalUrl = function(rawExternalUrlString: string) {
				$scope.loadingExternalUrlData = true;

				$timeout(function() {
					if ($scope.loadingExternalUrlData) {
						$scope.showExternalUrlProgressBar = true;
					}
				}, 1000);

				let successFunc = function(data) {
					$scope.externalUrlIsDupe = data.result.isDupe;
					$scope.externalUrlOriginalPageID = data.result.originalPageID;

					if (!!data.result.title) {
						$scope.submission.title = data.result.title;
					}

					$scope.showExternalUrlProgressBar = false;
					$scope.loadingExternalUrlData = false;
				};

				let failureFunc = function(data) {
					$scope.showExternalUrlProgressBar = false;
					$scope.loadingExternalUrlData = false;
				};

				arb.stateService.postData('/getExternalUrlData/', {rawExternalUrlString: rawExternalUrlString}, successFunc, failureFunc);
			};

			// Submit a new link to the feed.
			$scope.submitLink = function() {
				arb.stateService.postData('/newFeedPage/', $scope.submission, function(data) {
					$scope.feedRows.unshift(data.result.newFeedRow);
				});
				$scope.resetSubmission();
			};
		},
	};
});
