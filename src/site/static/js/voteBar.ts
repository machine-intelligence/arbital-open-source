import app from './angular.ts';

// Directive for showing a vote bar.
app.directive('arbVoteBar', function($http, $compile, $timeout, $mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/voteBar.html'),
		scope: {
			pageId: '@',
			isEmbedded: '=',
		},
		link: function(scope: any, element, attrs) {
			scope.arb = arb;
			scope.page = arb.stateService.pageMap[scope.pageId];
			scope.isHovering = false;
			scope.newVoteValue = undefined;
			// Bucket the user is hovering over (undefined is none, -1 is 'mu')
			scope.selectedVoteBucketIndex = undefined;

			// Return the vote object corresponding to current user's vote
			scope.getCurrentUserVote = function() {
				for (var i = 0; i < scope.page.votes.length; i++) {
					var vote = scope.page.votes[i];
					if (vote.userId === arb.userService.user.id) {
						return vote;
					}
				}
				return undefined;
			};

			// Set the value of current user's vote
			scope.setCurrentUserVote = function(voteValue) {
				arb.signupService.wrapInSignupFlow('vote', function() {
					var currentVote = scope.getCurrentUserVote();
					if (currentVote) {
						currentVote.value = voteValue;
					} else {
						scope.page.votes.push({
							value: voteValue,
							userId: arb.userService.user.id,
							createdAt: moment.utc().format('YYYY-MM-DD HH:mm:ss'),
						});
					}
					postNewVote();
				});
			};

			// Value of the current user's vote
			var typeHelpers = {
				probability: {
					headerLabel: 'Probability this claim is true',
					label1: '0%',
					label2: '25%',
					label3: '50%',
					label4: '75%',
					label5: '100%',
					toString: function(value) { return value == -1 ? '' : value + '%'; },
					buckets: [0,1,2,3,4,5,6,7,8,9],
					min: 0,
					max: 100,
					makeValid: function(value) { return Math.max(1, Math.min(99, Math.round(value))); },
					getFlex: function(n) { return 10; },
					getBucketIndex: function(value) { return Math.floor(value / 10); },
				},
				approval: {
					headerLabel: 'Approval ratings',
					label1: 'Strongly\ndisagree',
					label2: 'Disagree',
					label3: 'Neutral',
					label4: 'Agree',
					label5: 'Strongly\nagree',
					toString: function(value) {
						return '';
					},
					buckets: [0,1,2,3,4,5,6,7,8],
					min: 0,
					max: 100,
					makeValid: function(value) { return Math.max(0, Math.min(100, Math.round(value))); },
					getFlex: function(n) { return n == 4 ? 20 : 10; },
					getBucketIndex: function(value) {
						if (value <= -1) return -1;
						if (value >= 100) return 8;
						if (value <= 40) return Math.floor((value - 1) / 10);
						if (value >= 60) return Math.floor(value / 10) - 1;
						return 4;
					},
				},
			};
			scope.isProbability = scope.page.voteType === 'probability';
			scope.isApproval = scope.page.voteType === 'approval';
			scope.typeHelper = typeHelpers[scope.page.voteType];

			// Return normalized value for the votes in the given bucket [0.0-1.0]
			scope.getNormValue = function(bucketIndex) {
				var voteCount = 0;
				for (var i = 0; i < scope.page.votes.length; i++) {
					var vote = scope.page.votes[i];
					if (scope.typeHelper.getBucketIndex(vote.value) == bucketIndex) {
						voteCount++;
					}
				}
				return voteCount / scope.page.votes.length;
			};

			// Return a list of votes in the selected bucket.
			scope.getSelectedVotes = function() {
				if (scope.selectedVoteBucketIndex === undefined) return [];

				var votes = [];
				for (var i = 0; i < scope.page.votes.length; i++) {
					var vote = scope.page.votes[i];
					if (scope.typeHelper.getBucketIndex(vote.value) == scope.selectedVoteBucketIndex) {
						votes.push(vote);
					}
				}

				votes.sort(function(a, b) {
					if (a.value === b.value) {
						// Sort more recent votes first.
						return -a.createdAt.localeCompare(b.createdAt);
					}
					return a.value - b.value;
				});
				return votes;
			};

			// Send a new probability vote value to the server.
			var postNewVote = function() {
				var data = {
					pageId: scope.page.pageId,
					value: scope.getCurrentUserVote().value,
				};
				$http({method: 'POST', url: '/newVote/', data: JSON.stringify(data)})
				.error(function(data, status) {
					console.error('Error changing a vote:'); console.log(data); console.log(status);
				});
			};

			var $voteBarBody = element.find('.vote-bar-body');
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
				return value + '%';
			};

			// Hande mouse events
			scope.voteMouseMove = function(event, leave, muVote) {
				if (muVote) {
					scope.newVoteValue = undefined;
					scope.selectedVoteBucketIndex = leave ? undefined : -1;
					scope.isHovering = !leave;
					return;
				}
				scope.newVoteValue = scope.offsetToValue(event.pageX);
				if (!leave && event.pageY <= $voteBarBody.offset().top + $voteBarBody.height()) {
					scope.selectedVoteBucketIndex = scope.typeHelper.getBucketIndex(scope.newVoteValue);
				}
				if (leave && scope.getSelectedVotes().length <= 0) {
					scope.selectedVoteBucketIndex = undefined;
				}
				scope.isHovering = !leave;
			};
			scope.voteMouseClick = function(event) {
				scope.setCurrentUserVote(scope.offsetToValue(event.pageX));
			};
			// Called when the user casts a "mu" vote
			scope.muVote = function() {
				scope.setCurrentUserVote(-1);
			};

			// Return number of "mu" votes
			scope.getMuVoteCount = function() {
				var count = 0;
				for (var i = 0; i < scope.page.votes.length; i++) {
					if (scope.page.votes[i].value == -1) count++;
				}
				return count;
			};

			// Process deleting user's vote
			scope.deleteMyVote = function() {
				// TODO
				// scope.userVoteValue = undefined;
				postNewVote();
			};
		},
	};
});
