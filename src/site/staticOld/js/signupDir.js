"use strict";

// Directive for the Signup page.
app.directive("arbSignup", function(pageService, userService, $location) {
	return {
		templateUrl: "/static/html/signupDir.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;

			$(".submit-form").on("click", function(event) {
				var $form = $(".signup-form");
				var data = {};
				serializeFormData($form, data);
				submitForm($form, "/signup/", data, function(r) {
					window.location.href = $location.search().continueUrl;
				}, function() {
					console.log("ERROR");
				});
			});
		},
	};
});
