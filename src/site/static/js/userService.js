"use strict";

// User service.
app.service("userService", function($http, $location){
	var that = this;

	// Logged in user.
	this.user = undefined;

	// Map of all user objects.
	this.userMap = {};

	// Check if we can let this user do stuff.
	this.userIsCool = function() {
		return this.user && this.user.karma >= 200;
	};

	// Return url to the user page.
	this.getUserUrl = function(userId) {
		return "/user/" + userId;
	};

	// Return a user's full name.
	this.getFullName = function(userId) {
		var user = this.userMap[userId];
		if (!user) console.error("User not found: " + userId);
		return user.firstName + " " + user.lastName;
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
	};

	this.isTouchDevice = "ontouchstart" in window // works in most browsers
		|| (navigator.MaxTouchPoints > 0)
		|| (navigator.msMaxTouchPoints > 0);

	// Sign into FB and call the callback with the response.
	this.fbLogin = function(callback) {
		// Apparently FB.login is not supported in Chrome in iOS
		if (true || navigator.userAgent.match("CriOS")) {
			var appId = isLive() ? "1064531780272247" : "1064555696936522";
			var redirectUrl = encodeURIComponent(isLive() ? "http://arbital.com/" : "http://localhost:8012/");
			window.location.href = "https://www.facebook.com/dialog/oauth?client_id=" + appId +
					"&redirect_uri=" + redirectUrl + "&scope=email,public_profile";
		} else {
			FB.login(function(response){
				callback(response);
			}, {scope: 'public_profile,email'});
		}
	};

	// Check if FB redirected back to use with the code
	if ($location.search().code) {
		var data = {
			fbCodeToken: $location.search().code,
		};
		$http({method: "POST", url: "/signup/", data: JSON.stringify(data)})
		.success(function(data, status){
			window.location.href = $location.search().continueUrl || "/";
		})
		.error(function(data, status){
			console.error("Error FB signup:"); console.log(data); console.log(status);
		});
	}
});
