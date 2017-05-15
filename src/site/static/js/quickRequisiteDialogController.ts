import app from './angular.ts';

// EditClaimDialogController is used for editing a claim in an mdDialog
app.controller('QuickRequisiteDialogController', function($scope, $mdDialog, $timeout, $interval, $http, arb, originalPage) {
	$scope.arb = arb;

	let focusInput = function() {
		$('#edit-claim-title-input').focus();
	};

	// Called when a user presses a key inside the title input
	$scope.titleKeypress = function(event) {
		if (event.charCode == 13) {
			$scope.publishPage();
		}
	};

	// Called when the user is done editing the page
	$scope.doneFn = function(result) {
		$mdDialog.hide(result);
	};

	// Called when the user closed the dialog
	$scope.hide = function() {
		$mdDialog.hide({hidden: true, pageId: $scope.pageId});
	};

	// Called when user clicks Publish button
	$scope.publishPage = function() {
		arb.pageService.savePageInfo($scope.page, function() {
			var data = arb.editService.computeSavePageData($scope.page);
			$http({method: 'POST', url: '/editPage/', data: JSON.stringify(data)})
				.success(function(returnedData) {
					$scope.doneFn({pageId: $scope.page.pageId});
				})
				.error(function(returnedData) {
					arb.popupService.showToast({text: 'Error saving the claim.', isError: true});
				});
		});
	};

	var searchingForSimilarPages = false;
	var prevSimilarData = {};
	$scope.similarPages = [];
	var findSimilarFunc = function() {
		if (searchingForSimilarPages) return;
		if (!$scope.page) return;

		var data = {
			title: $scope.page.title,
			onlyClaims: true,
		};
		if (JSON.stringify(data) == JSON.stringify(prevSimilarData)) {
			return;
		}

		searchingForSimilarPages = true;
		prevSimilarData = data;
		arb.autocompleteService.findSimilarPages(data, function(data) {
			searchingForSimilarPages = false;
			$scope.similarPages.length = 0;
			for (var n = 0; n < data.length; n++) {
				var pageId = data[n].pageId;
				if (pageId === $scope.page.pageId) continue;
				$scope.similarPages.push({pageId: pageId, score: data[n].score});
			}
		});
	};
	var similarInterval = $interval(findSimilarFunc, 500);
	$scope.$on('$destroy', function() {
		$interval.cancel(similarInterval);
	});

	$timeout(function() {
		// We need this class in the beginning to make sure dialog appears in the
		// correct position; but then we need to remove it, otherwise the body
		// is not visible.
		$('body').removeClass('md-dialog-is-showing');
	});
});
