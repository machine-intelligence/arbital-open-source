// The code for handling likes-related stuff
'use strict';

app.service('likeService', function($mdDialog, signupService, stateService) {

	// Report a like click
	this.processLikeClick = function(likeable, objectId, value) {
		signupService.wrapInSignupFlow('like', function() {
			if (!likeable) return;
			if (value) {
				likeable.myLikeValue = value;
			} else {
				likeable.myLikeValue = Math.min(1, 1 - likeable.myLikeValue);
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