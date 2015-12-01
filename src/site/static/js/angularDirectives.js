"use strict";

// userName directive displayes a user's name.
app.directive("arbUserName", function(userService) {
	return {
		templateUrl: "/static/html/userName.html",
		scope: {
			userId: "@",
		},
		link: function(scope, element, attrs) {
			scope.userService = userService;
			scope.user = userService.userMap[scope.userId];
		},
	};
});

// newLinkModal directive is used for storing a modal that creates new links.
app.directive("arbNewLinkModal", function(autocompleteService) {
	return {
		templateUrl: "/static/html/newLinkModal.html",
		scope: {
		},
		link: function(scope, element, attrs) {
			var $input = element.find(".new-link-input");
			// Set up autocomplete
			/*autocompleteService.setupParentsAutocomplete($input, function(event, ui) {
				element.find(".modal-content").submit();
				return true;
			});*/
			// Set up search for new link modal
			//autocompleteService.setAutocompleteRendering($input, scope);
		},
	};
});

// intrasitePopover containts the popover body html.
app.directive("arbIntrasitePopover", function(pageService, userService) {
	return {
		templateUrl: "/static/html/intrasitePopover.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on("$destroy", function() {
				isDestroyed = true;
			});

			// Add the primary page as the first lens.
			if (scope.page.lensIds.indexOf(scope.page.pageId) < 0) {
				scope.page.lensIds.unshift(scope.page.pageId);
			}

			// Check if a lens is loaded
			scope.isLoaded = function(lensId) {
				return pageService.pageMap[lensId].summary.length > 0;
			};

			// Called when a tab is selected
			scope.tabSelect = function(lensId) {
				if (scope.isLoaded(lensId)) return;
				// Fetch page data from the server.
				pageService.loadIntrasitePopover(lensId, {
					success: function() {
						if (isDestroyed) return;
						var lens = pageService.pageMap[lensId];
						if (!lens.summary) {
							lens.summary = " "; // to avoid trying to load it again
						}
						// Page's lensIds got resent, so need to fix this again
						if (scope.page.lensIds.indexOf(scope.page.pageId) < 0) {
							scope.page.lensIds.unshift(scope.page.pageId);
						}
					},
				});
			};
		},
	};
});

// pageTitle displays page's title with optional meta info.
app.directive("arbPageTitle", function(pageService, userService) {
	return {
		templateUrl: "/static/html/pageTitle.html",
		scope: {
			pageId: "@",
			// Options override for the page's title
			customPageTitle: "@",
			// Whether to display the title as a link or a span
			isLink: "@",
			// Whether or not to show the clickbait
			showClickbait: "@",
			// Whether or not to show the type of the page icon
			showType: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
			scope.pageTitle = scope.page.title;
			if (scope.customPageTitle) {
				scope.pageTitle = scope.customPageTitle;
			}
		},
	};
});

// confirmPopover displays a confirmation popover, with a custom message,
// with callbacks for confirm and cancel, which get passed pageId
app.directive("arbConfirmPopover", function(pageService, userService) {
	return {
		templateUrl: "/static/html/confirmPopover.html",
		scope: {
			message: "@",
			pageId: "@",
			xPos: "@",
			yPos: "@",
			// The callbacks will close the popover if the return value is not true
			confirmFn: "&",
			// The cancel callback is optional.  If there is no cancel callback, the popover will simply close
			// If this is not set, then angular will use an empty function, that returns "undefined"
			cancelFn: "&",
		},
		link: function(scope, element, attrs) {
			element.find(".confirm-popover-button").on("click", function(event) {
				var result = scope.confirmFn({returnedPageId: scope.pageId});
				if (!result) {
					element.remove();
				}
			});
			element.find(".cancel-popover-button").on("click", function(event) {
				var result = scope.cancelFn({returnedPageId: scope.pageId});
				if (!result) {
					element.remove();
				}
			});
		},
	};
});

// likes displays the likes button(s) for a page.
app.directive("arbLikes", function($http, pageService, userService) {
	return {
		templateUrl: "/static/html/likes.html",
		scope: {
			pageId: "@",
			// If true, the button is not an icon button, but is a normal button
			isStretched: "@",
			// Whether or not we show likes as a button or a span
			isButton: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];

			// User clicked on the like button
			scope.likeClick = function() {
				scope.page.myLikeValue = 1 - scope.page.myLikeValue;

				var data = {
					pageId: scope.page.pageId,
					value: scope.page.myLikeValue,
				};
				$http({method: "POST", url: "/newLike/", data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error changing a like:"); console.log(data); console.log(status);
				});
			};
		},
	};
});

// subscribe directive displays the button for subscribing to a page.
app.directive("arbSubscribe", function($http, pageService, userService) {
	return {
		templateUrl: "/static/html/subscribe.html",
		scope: {
			pageId: "@",
			// If true, the button is not an icon button, but is a normal button with a label
			isStretched: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];

			// User clicked on the subscribe button
			scope.subscribeClick = function() {
				scope.page.isSubscribed = !scope.page.isSubscribed;
				var data = {
					pageId: scope.page.pageId,
				};
				var url = scope.page.isSubscribed ? "/newSubscription/" : "/deleteSubscription/";
				$http({method: "POST", url: url, data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error changing a subscription:"); console.log(data); console.log(status);
				});
			};
		},
	};
});

// Directive for showing a vote bar.
app.directive("arbVoteBar", function($http, $compile, $timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/voteBar.html",
		scope: {
			pageId: "@",
			isPopoverVote: "@",
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
					headerLabel: "What's the probability of this claim being true?",
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
					headerLabel: "What's the approval rating of this proposal?",
					label1: "Strongly\nDisapprove",
					label2: "Disapprove",
					label3: "Neutral",
					label4: "Approve",
					label5: "Strongly\nApprove",
					toString: function(value) {
						return (value > 0 ? "+" : "") + (value / 10).toFixed(1);
					},
					bucketCount: 9,
					min: -50,
					max: 50,
					makeValid: function(value) { return Math.max(-50, Math.min(50, Math.round(value))); },
					getFlex: function(n) { return n == 4 ? 20 : 10; },
					getBucketIndex: function(value) {
						value = (value < 0 ? value + 1 : value - 1) / 10;
						value = value < 0 ? Math.ceil(value) : Math.floor(value);
						return value + 4;
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

// composeFab is the FAB button in the bottom right corner used for creating new pages
app.directive("arbComposeFab", function($location, pageService, userService) {
	return {
		templateUrl: "/static/html/composeFab.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.$on("$locationChangeSuccess", function () {
				$scope.hide = $location.path().indexOf("/edit") === 0;
			});
			$scope.hide = $location.path().indexOf("/edit") === 0;

			$scope.newPage = function() {
				$location.url("/edit");
			}

			$scope.newSibling = function() {
				var parentIds = pageService.primaryPage.parentIds.join(",");
				$location.url("/edit?newParentId=" + parentIds);
			}

			$scope.newChild = function() {
				$location.url("/edit?newParentId=" + pageService.primaryPage.pageId);
			}
		},
	};
});

// autocomplete searches for relevant pages as you do the search
app.directive("arbAutocomplete", function($q, pageService, autocompleteService) {
	return {
		templateUrl: "/static/html/autocomplete.html",
		scope: {
			autofocus: "@",
			placeholder: "@",
			onSelect: "&",
		},
		controller: function($scope) {
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				autocompleteService.performSearch({term: text}, function(results) {
					deferred.resolve(results);
				});
        return deferred.promise;
			};

			$scope.searchResultSelected = function(result) {
				$scope.onSelect({result: result});
				$scope.searchText = "";
			};
		},
	};
});

