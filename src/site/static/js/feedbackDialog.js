// FeedbackDialogController is used for submitting feedback via mdDialog.
app.controller('FeedbackDialogController', function($scope, $mdDialog, $timeout, $http, userService, pageService) {
	$scope.arb = arb;
	

	// Submit feedback
	$scope.submitFn = function() {
		var data = {
			text: $scope.text,
		};
		$http({method: 'POST', url: '/feedback/', data: JSON.stringify(data)});
		$mdDialog.hide();
	};

	// Called when the user closed the dialog
	$scope.hide = function() {
		$mdDialog.hide();
	};

	$timeout(function() {
		$('.feedback-textarea').focus();
		// We need this class in the beginning to make sure dialog appears in the
		// correct position; but then we need to remove it, otherwise the body
		// is not visible.
		$('body').removeClass('md-dialog-is-showing');
	});
});
