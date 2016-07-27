// The code for the signup flow
'use strict';

import app from './angular.ts';

app.service('signupService', function($mdDialog, $timeout, analyticsService, userService, stateService) {
	var that = this;

	// The function that we call when the signup is finished.
	that.afterSignupFn = undefined;
	// The string name of the action that was attempted that triggered the signup.
	that.attemptedAction = undefined;

	// Trigger the signup flow
	// - attemptedAction: the string name of the action that triggered the signup flow
	// - afterSignupFn: the function to call if signup succeeds
	that.wrapInSignupFlow = function(attemptedAction, afterSignupFn) {
		if (userService.userIsLoggedIn()) {
			afterSignupFn();
			return;
		}

		that.afterSignupFn = afterSignupFn;
		that.attemptedAction = attemptedAction;
		that.openSignupDialog();
	};

	// Open the signup dialog
	that.openSignupDialog = function() {
		$mdDialog.show({
			template: '<arb-signup></arb-signup>',
			clickOutsideToClose: true,
		});
		$timeout(function() {
			$('body').removeClass('md-dialog-is-showing');
		});
		analyticsService.reportSignupAction('view signup form', that.attemptedAction);
	};

	// Close the signup dialog
	that.closeSignupDialog = function() {
		$mdDialog.hide();

		if (that.afterSignupFn) that.afterSignupFn();
	};

	// TODO: Roggie asks: why is this in signupService.js? Why not define the function where it's used and just call signupService.wrapInSignupFlow from there?
	// Report a like click
	that.processLikeClick = function(likeable, objectId, value) {
		that.wrapInSignupFlow('like', function() {
			if (!likeable) return;
			if (value) {
				likeable.myLikeValue = value;
			} else if (likeable.myLikeValue == 0) {
				likeable.myLikeValue = 1;
			} else {
				likeable.myLikeValue = 0;
			}
			var data = {
				likeableId: likeable.likeableId,
				objectId: objectId,
				likeableType: likeable.likeableType,
				value: likeable.myLikeValue,
			};
			stateService.postDataWithoutProcessing('/newLike/', data);
		});
	};
});
