"use strict";

// Set up triggers.
$(function() {
	// Process new member form submission.
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

	// Process new group form submission.
	var $form = $("#new-group-form");
	$form.on("submit", function(event) {
		var data = {
			name: $form.attr("name"),
		};
		submitForm($form, "/newGroup/", data, function(r) {
			location.reload();
		});
		return false;
	});
});
