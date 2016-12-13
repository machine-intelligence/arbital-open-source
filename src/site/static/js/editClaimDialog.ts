import app from './angular.ts';

// EditClaimDialogController is used for editing a claim in an mdDialog
app.controller('EditClaimDialogController', function($scope, $mdDialog, $timeout, $http, arb, originalPageId, title, resumePageId) {
	$scope.arb = arb;

	let focusInput = function() {
		$('#edit-claim-title-input').focus();
	};

	// Create new page
	if (!resumePageId) {
		arb.pageService.getNewPage({
			type: 'wiki',
			parentIds: [originalPageId],
			success: function(newPageId) {
				$scope.pageId = newPageId;
				$scope.page = arb.stateService.editMap[$scope.pageId];
				$scope.page.hasVote = true;
				$scope.page.voteType = 'approval';
				$scope.page.text = ' ';
				$scope.page.seeDomainId = arb.stateService.editMap[originalPageId].seeDomainId;
				$scope.page.editDomainId = arb.stateService.editMap[originalPageId].editDomainId;
				if (title) {
					$scope.page.title = title;
				}
				$timeout(focusInput);
			},
		});
	} else {
		arb.pageService.loadEdit({
			pageAlias: resumePageId,
			success: function() {
				$scope.pageId = resumePageId;
				$scope.page = arb.stateService.editMap[$scope.pageId];
				if (!$scope.page.voteType) {
					$scope.page.voteType = arb.editService.voteTypes.approval;
				}
				$timeout(focusInput);
			},
		});
	}

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

	$timeout(function() {
		// We need this class in the beginning to make sure dialog appears in the
		// correct position; but then we need to remove it, otherwise the body
		// is not visible.
		$('body').removeClass('md-dialog-is-showing');
	});
});
