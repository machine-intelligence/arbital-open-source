"use strict";

// Set up triggers.
$(function() {
	// Process new domain form submission.
	var $form = $("#new-domain-form");
	$form.on("submit", function(event) {
		var data = {
			name: $form.attr("name"),
			alias: $form.attr("alias"),
			rootPageId: $form.attr("rootPageId"),
			isDomain: true,
		};
		submitForm($form, "/newGroup/", data, function(r) {
			location.reload();
		});
		return false;
	});
});
