// Directive for showing a vote bar.
app.directive("arbVoteBar", function($http, $compile, $timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/voteBar.html",
		scope: {
			pageId: "@",
			isEmbedded: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			var userId = userService.user.id;

			// Value of the current user's vote
			scope.userVoteValue = undefined;
			var typeHelpers = {
				probability: {
					headerLabel: "Probability of this claim being true",
					label1: "0%",
					label2: "25%",
					label3: "50%",
					label4: "75%",
					label5: "100%",
					toString: function(value) { return value + "%"; },
					bucketCount: 10,
					min: 0,
					max: 100,
					makeValid: function(value) { return Math.max(1, Math.min(99, Math.round(value))); },
					getFlex: function(n) { return 10; },
					getBucketIndex: function(value) { return Math.floor(value / 10); },
				},
				approval: {
					headerLabel: "Approval ratings",
					label1: "Strongly\nDisapprove",
					label2: "Disapprove",
					label3: "Neutral",
					label4: "Approve",
					label5: "Strongly\nApprove",
					toString: function(value) {
						return "";
					},
					bucketCount: 9,
					min: 0,
					max: 100,
					makeValid: function(value) { return Math.max(0, Math.min(100, Math.round(value))); },
					getFlex: function(n) { return n == 4 ? 20 : 10; },
					getBucketIndex: function(value) {
						if (value <= 0) return 0;
						if (value >= 100) return 8;
						if (value <= 40) return Math.floor((value - 1) / 10);
						if (value >= 60) return Math.floor(value / 10) - 1;
						return 4;
					},
				},
			};
			scope.isProbability = scope.page.voteType === "probability";
			scope.isApproval = scope.page.voteType === "approval";
			scope.typeHelper = typeHelpers[scope.page.voteType];

			// Create vote buckets
			scope.voteBuckets = [];
			for (var n = 0; n < scope.typeHelper.bucketCount; n++) {
				scope.voteBuckets.push({normValue: 0, flex: scope.typeHelper.getFlex(n), votes: []});
			}
			// Fill buckets.
			for(var i = 0; i < scope.page.votes.length; i++) {
				var vote = scope.page.votes[i];
				var bucket = scope.voteBuckets[scope.typeHelper.getBucketIndex(vote.value)];
				if (vote.userId === userService.user.id) {
					scope.userVoteValue = vote.value;
				} else {
					bucket.votes.push({userId: vote.userId, value: vote.value, createdAt: vote.createdAt});
				}
			}
			// Normalize values and sort votes.
			for (var n = 0; n < scope.typeHelper.bucketCount; n++) {
				scope.voteBuckets[n].normValue = scope.voteBuckets[n].votes.length / scope.page.votes.length;
				scope.voteBuckets[n].votes.sort(function(a, b) {
					if (a.value === b.value) {
						return a.createdAt < b.createdAt;
					}
					return a.value - b.value;
				});
			}

			// Send a new probability vote value to the server.
			var postNewVote = function() {
				var data = {
					pageId: scope.page.pageId,
					value: scope.userVoteValue || 0.0,
				};
				$http({method: "POST", url: "/newVote/", data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error changing a vote:"); console.log(data); console.log(status);
				});
			}

			var $voteBarBody = element.find(".vote-bar-body");
			// Bucket the user is hovering over
			scope.selectedVoteBucket = undefined;
			// Convert mouseX position to selected value on the bar.
			scope.offsetToValue = function(pageX) {
				var range = scope.typeHelper.max - scope.typeHelper.min;
				var value = ((pageX - $voteBarBody.offset().left) * range) / $voteBarBody.width() + scope.typeHelper.min;
				return scope.typeHelper.makeValid(value);
			};
			// Convert given value to 0-100% offset for the bar.
			scope.valueToOffset = function(value) {
				var range = scope.typeHelper.max - scope.typeHelper.min;
				value = ((value - scope.typeHelper.min) * 100) / range;
				return value + "%";
			};

			// Hande mouse events
			scope.isHovering = false;
			scope.newVoteValue = undefined;
			scope.voteMouseMove = function(event, leave) {
				scope.newVoteValue = scope.offsetToValue(event.pageX);
				scope.selectedVoteBucket = scope.voteBuckets[scope.typeHelper.getBucketIndex(scope.newVoteValue)];
				if (leave && scope.selectedVoteBucket.votes.length <= 0) {
					scope.selectedVoteBucket = undefined;
				}
				scope.isHovering = !leave;
			};
			scope.voteMouseClick = function(event, leave) {
				scope.userVoteValue = scope.offsetToValue(event.pageX);
				postNewVote();
			};

			// Process deleting user's vote
			scope.deleteMyVote = function() {
				scope.userVoteValue = undefined;
				postNewVote();
			};
		},
	};
});
