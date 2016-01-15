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

	// =========== Questionnaire helpers ====================
	// Map questionIndex -> {answer: 'a', knows: [ids], wants: [ids]}
	this.answers = {};
	this.setQuestionAnswer = function(qIndex, answer, aKnows, aWants) {
		// Compute existing knows and wants
		var knows = [], wants = [];
		for (var i in this.answers) {
			knows = knows.concat(this.answers[i].knows);
			wants = wants.concat(this.answers[i].wants);
		};
		this.answers[qIndex] = {answer: answer, knows: aKnows, wants: aWants};
		// Compute updated knows and wants
		var newKnows = [], newWants = [];
		for (var a in this.answers) {
			newKnows = newKnows.concat(this.answers[a].knows);
			newWants = newWants.concat(this.answers[a].wants);
		};
		// Compute the diff and update the BE
		for (var n = 0; n < newKnows.length; n++) {
			if (knows.indexOf(newKnows[n]) < 0) {
				//added know
			}
		}
		for (var n = 0; n < newWants.length; n++) {
			if (knows.indexOf(newWants[n]) < 0) {
				//added want
			}
		}
		for (var n = 0; n < knows.length; n++) {
			if (newKnows.indexOf(knows[n]) < 0) {
				//removed know
			}
		}
		for (var n = 0; n < wants.length; n++) {
			if (newWants.indexOf(wants[n]) < 0) {
				//removed want
			}
		}
		// Update the answers on the BE?
	};
});
