"use strict";

// Set up triggers.
$(function() {
	// Process new member form submission.
	$(".new-member-form").on("submit", function(event) {
		var $target = $(event.target);
		var data = {
			groupName: $target.attr("group-name"),
		};
		submitForm($target, "/newMember/", data, function(r) {
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
