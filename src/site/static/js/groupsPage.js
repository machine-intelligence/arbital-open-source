"use strict";

// Set up triggers.
$(function() {
	// Process form submission.
	var $form = $("#new-member-form");
	$form.on("submit", function(event) {
		var data = {
			groupName: $form.attr("group-name"),
		};
		submitForm($form, "/newMember/", data, function(r) {
			location.reload();
		});
		return false;
	});
});
