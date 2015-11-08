"use strict";

// User service.
app.service("userService", function(){
	var that = this;

	// Logged in user.
	this.user = undefined;

	// Map of all user objects.
	this.userMap = {};

	// Check if we can let this user do stuff.
	this.userIsCool = function() {
		return this.user.karma >= 200;
	};

	// Return url to the user page.
	this.getUserUrl = function(userId) {
		return "/user/" + userId;
	};

	// (Un)subscribe a user to a thing.
	var subscribeTo = function(doSubscribe, data, done) {
		$.ajax({
			type: "POST",
			url: doSubscribe ? "/newSubscription/" : "/deleteSubscription/",
			data: JSON.stringify(data),
		})
		.done(done);
	};
	// (Un)subscribe a user to another user.
	this.subscribeToUser = function($target) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			userId: $target.attr("user-id"),
		};
		subscribeTo($target.hasClass("on"), data, function(r) {});
	}
	this.subscribeToPage = function($target) {
		var $target = $(event.target);
		$target.toggleClass("on");
		var data = {
			pageId: $target.attr("page-id"),
		};
		subscribeTo($target.hasClass("on"), data, function(r) {});
	};

	// Call this to process data we received from the server.
	this.processServerData = function(data) {
		if (data.resetEverything) {
			that.userMap = {};
		}
		if (data.user) {
			that.user = data.user;
		}
		$.extend(that.userMap, data["users"]);
	}
});
