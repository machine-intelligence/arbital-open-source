"use strict";

// Setup handlers.
$(function() {
	// Subscription stuff.
	$(".subscribe-to-user-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			userId: $target.attr("user-id"),
		};
		$.ajax({
			type: 'POST',
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
	$(".subscribe-to-tag-link").on("click", function(event) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			tagId: $target.attr("tag-id"),
		};
		$.ajax({
			type: 'POST',
			url: $target.hasClass("on") ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(function(r) {
		});
		return false;
	});
});
