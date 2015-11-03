$(function() {
	$(".submit-form").on("click", function(event) {
		var $form = $(".signup-form");
		var data = {};
		serializeFormData($form, data);
		submitForm($form, "/signup/", data, function(r) {
			window.location.href = r;
		}, function() {
			console.log("ERROR");
		});
	});
});
