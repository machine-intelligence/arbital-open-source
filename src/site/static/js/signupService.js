// The code for the signup flow
'use strict';

app.service('signupService', function($mdDialog, analyticsService) {
	var that = this;

	// Open the signup dialog
	that.openSignupDialog = function(opt_attemptedAction) {
		that.attemptedAction = opt_attemptedAction;

		// Show a signup dialog over the page
		$mdDialog.show({
			template: '<arb-signup></arb-signup>',
			clickOutsideToClose: true,
		});
		analyticsService.reportSignupAction('view signup form', that.attemptedAction);
	};

	// Close the signup dialog
	that.closeSignupDialog = function() {
		$mdDialog.hide();
	};
});