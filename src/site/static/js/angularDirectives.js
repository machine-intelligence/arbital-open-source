"use strict";

// userName directive displayes a user's name.
app.directive("arbUserName", function(userService) {
	return {
		templateUrl: "/static/html/userName.html",
		scope: {
			userId: "@",
		},
		controller: function($scope) {
			$scope.userService = userService;
			$scope.user = userService.userMap[$scope.userId];
		},
	};
});

// intrasitePopover contains the popover body html.
app.directive("arbIntrasitePopover", function($timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/intrasitePopover.html",
		scope: {
			pageId: "@",
			direction: "@",
			arrowOffset: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];
			$scope.summaries = [];
			$scope.getArrowStyle = function() {
				return {"left": +$scope.arrowOffset};
			};

			// Check if summaries are loaded
			$scope.isLoaded = function() {
				return $scope.summaries.length > 0;
			};
		},
		link: function(scope, element, attrs) {
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on("$destroy", function() {
				isDestroyed = true;
			});

			// Convert the name of the tab into an index for sorting.
			var nameToTabIndex = function(name) {
				if (name === "Brief") return 0;
				if (name === "Summary") return 1;
				if (name === "Technical") return 2;
				return 3;
			};

			// Convert page's summaries into our local array
			var processPageSummaries = function() {
				for (var name in scope.page.summaries) {
					scope.summaries.push({name: name, text: scope.page.summaries[name]});
				}
				scope.summaries.sort(function(a, b) {
					return nameToTabIndex(a.name) > nameToTabIndex(b.name);
				});
			};

			processPageSummaries();
			if (scope.summaries.length <= 0) {
				// Fetch page summaries from the server.
				pageService.loadIntrasitePopover(scope.pageId, {
					success: function() {
						if (isDestroyed) return;
						processPageSummaries();
						// Hack: we need to fix the md-tabs height, because it takes way too long
						// to adjust by itself.
						$timeout(function() {
							var $el = element.find(".popover-tab-body");
							$el.closest("md-tabs").height($el.children().height());
						});
					},
				});
			}
		},
	};
});

// userPopover contains the popover body html.
app.directive("arbUserPopover", function($timeout, pageService, userService) {
	return {
		templateUrl: "/static/html/userPopover.html",
		scope: {
			userId: "@",
			direction: "@",
			arrowOffset: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.user = userService.userMap[$scope.userId];
			$scope.page = pageService.pageMap[$scope.userId];
			$scope.summaries = [];

			$scope.getArrowStyle = function() {
				return {"left": +$scope.arrowOffset};
			};

			// Check if the data is loaded
			$scope.isLoaded = function() {
				if ($scope.user) {
					return $scope.summaries.length > 0;
				}
				return false;
			};
		},
		link: function(scope, element, attrs) {
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on("$destroy", function() {
				isDestroyed = true;
			});

			// Convert page's summaries into our local array
			var processPageSummaries = function() {
				if (!scope.page || !scope.page.summaries) return;
				for (var name in scope.page.summaries) {
					scope.summaries.push({name: name, text: scope.page.summaries[name]});
				}
			};

			processPageSummaries();
			if (!scope.isLoaded()) {
				pageService.loadUserPopover(scope.userId, {
					success: function() {
						if (isDestroyed) return;
						scope.user = userService.userMap[scope.userId];
						scope.page = pageService.pageMap[scope.userId];
						processPageSummaries();
						// Hack: we need to fix the md-tabs height, because it takes way too long
						// to adjust by itself.
						$timeout(function() {
							var $el = element.find(".popover-tab-body");
							$el.closest("md-tabs").height($el.children().height());
						});
					},
				});
			}
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
			customPageTitle: "=",
			// Whether to display the title as a link or a span
			isLink: "@",
			// Whether or not to show the clickbait
			showClickbait: "@",
			// Whether or not to show the type of the page icon
			showType: "@",
			// If set, we'll pull the page from the edit map
			useEditMap: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = ($scope.useEditMap ? pageService.editMap : pageService.pageMap)[$scope.pageId];

			$scope.getTitle = function() {
				if ($scope.customPageTitle) {
					return $scope.customPageTitle;
				}
				if ($scope.page.isComment()) {
					return "*Comment*";
				}
				if ($scope.page.isAnswer() && !$scope.page.title) {
					return "*Answer*";
				}
				return $scope.page.title;
			};
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
			isStretched: "=",
			// Whether or not we show likes as a button or a span
			isButton: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			// User clicked on the like button
			$scope.likeClick = function() {
				$scope.page.myLikeValue = Math.min(1, 1 - $scope.page.myLikeValue);

				var data = {
					pageId: $scope.page.pageId,
					value: $scope.page.myLikeValue,
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
			// If true, subscribe to a user, not a page
			isUser: "=",
			// If true, the button is not an icon button, but is a normal button with a label
			isStretched: "=",
			showSubscriberCount: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			// Check if the data is loaded
			$scope.isSubscribed = function() {
				if (!$scope.isUser) {
					return pageService.pageMap[$scope.pageId].isSubscribed;
				} else {
					return userService.userMap[$scope.pageId].isSubscribed;
				}
			};

			// User clicked on the subscribe button
			$scope.subscribeClick = function() {
				if (!$scope.isUser) {
					pageService.pageMap[$scope.pageId].isSubscribed = !$scope.isSubscribed();
				} else {
					userService.userMap[$scope.pageId].isSubscribed = !$scope.isSubscribed();
				}
				var data = {
					pageId: $scope.pageId,
				};
				var url = $scope.isSubscribed() ? "/newSubscription/" : "/deleteSubscription/";
				$http({method: "POST", url: url, data: JSON.stringify(data)})
				.error(function(data, status){
					console.error("Error changing a subscription:"); console.log(data); console.log(status);
				});
			};
		},
	};
});

// composeFab is the FAB button in the bottom right corner used for creating new pages
app.directive("arbComposeFab", function($location, $timeout, $mdMedia, $mdDialog, pageService, userService) {
	return {
		templateUrl: "/static/html/composeFab.html",
		scope: {
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.pageUrl = "/edit/";
			$scope.isSmallScreen = !$mdMedia("gt-sm");

			var isTouchDevice = "ontouchstart" in window // works in most browsers
					|| (navigator.MaxTouchPoints > 0)
					|| (navigator.msMaxTouchPoints > 0);

			$scope.isOpen = false;
			$scope.toggle = function(show, hovering) {
				if (isTouchDevice) return;
				$scope.isOpen = show;
			};

			// Compute what the urls should be on the compose buttons, and which ones
			// should be visible.
			var computeUrls = function() {
				$scope.questionUrl = "/edit/?type=question";
				$scope.editPageUrl = undefined;
				$scope.childUrl = undefined;
				$scope.lensUrl = undefined;
				$scope.showNewAnswer = false;
				if (pageService.primaryPage) {
					var type = pageService.primaryPage.type;
					if (type === "question") {
						$scope.showNewAnswer = true;
					} else if (type === "wiki" || type === "group" || type === "domain") {
						$scope.questionUrl = "/edit/?newParentId=" + pageService.primaryPage.pageId + "&type=question";
						$scope.lensUrl = "/edit/?newParentId=" + pageService.primaryPage.pageId + "&type=lens";
						$scope.childUrl = "/edit?newParentId=" + pageService.primaryPage.pageId;
					}
					if ($location.search().lens) {
						$scope.editPageUrl = pageService.getEditPageUrl($location.search().lens);
					} else {
						$scope.editPageUrl = pageService.getEditPageUrl(pageService.primaryPage.pageId);
					}
				}
			};
			computeUrls();
			$scope.$watch(function() {
				// Note: can't use an object, so we just hack together a string
				return (pageService.primaryPage ? pageService.primaryPage.pageId : "none") + $location.absUrl();
			}, function() {
				computeUrls();
			});

			// New answer button is clicked
			$scope.newAnswer = function() {
				$("html, body").animate({
					scrollTop: $("#your-answer").offset().top,
		    }, 1000);
			};

			// New feedback button is clicked
			$scope.newFeedback = function(event) {
				$mdDialog.show({
					templateUrl: "/static/html/feedbackDialog.html",
					controller: "FeedbackDialogController",
					autoWrap: false,
					targetEvent: event,
				});
			};

			$scope.$on("$locationChangeSuccess", function () {
				$scope.hide = $location.path().indexOf("/edit") === 0;
			});
			$scope.hide = $location.path().indexOf("/edit") === 0;

			// Listen for shortcut keys
			$(document).keyup(function(event) {
				if (!event.ctrlKey || !event.altKey) return true;
				$scope.$apply(function() {
					if (event.keyCode == 80) $location.url("/edit/"); // P
					else if (event.keyCode == 69 && $scope.editPageUrl) $location.url($scope.editPageUrl); // E
					else if (event.keyCode == 67 && $scope.childUrl) $location.url($scope.childUrl); // C
					else if (event.keyCode == 78 && $scope.lensUrl) $location.url($scope.lensUrl); // N
					else if (event.keyCode == 81 && $scope.questionUrl) $location.url($scope.questionUrl); // Q
					else if (event.keyCode == 65 && $scope.showNewAnswer) $scope.newAnswer(); // A
					else if (event.keyCode == 75) $scope.newFeedback(event); // K
				});
			});
		},
	};
});

// autocomplete searches for relevant pages as you do the search
app.directive("arbAutocomplete", function($timeout, $q, pageService, userService, autocompleteService) {
	return {
		templateUrl: "/static/html/autocomplete.html",
		scope: {
			doAutofocus: "=",
			placeholder: "@",
			// If set, the search will be constrained to this page type
			pageType: "@",
			onSelect: "&",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				autocompleteService.performSearch({term: text, pageType: $scope.pageType}, function(results) {
					deferred.resolve(results);
				});
        return deferred.promise;
			};

			$scope.searchResultSelected = function(result) {
				$scope.onSelect({result: result});
				$scope.searchText = ""; // this triggers this function
			};
		},
	};
});

// confirmButton is a button that ask for a confirmation when you press it
app.directive("arbConfirmButton", function() {
	return {
		templateUrl: "/static/html/confirmButton.html",
		scope: {
			buttonText: "@",
			buttonBeforeConfirm: "@",
			disabled: "=",
			confirmed: "&",
		},
		controller: function($scope) {
			$scope.confirming = false;
			$scope.buttonFlexOrder = $scope.buttonBeforeConfirm ? -1 : 1;

			$scope.toggleConfirming = function(confirming) {
				$scope.confirming = confirming;
			};
		},
	};
});

// Directive for the User page panel
app.directive("arbPageList", function(pageService, userService) {
	return {
		templateUrl: "/static/html/pageList.html",
		scope: {
			pageIds: "=",
			panelTitle: "@",
			hideLikes: "=",
			showLastEdit: "=",
			showCreatedAt: "=",
			showQuickEdit: "=",
			showRedLinkCount: "=",
			showCommentCount: "=",
			showTextLength: "=",
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			$scope.getPage = function(pageId) {
				if ($scope.useEditMap) {
					return pageService.editMap[pageId];
				} 
				return pageService.pageMap[pageId];
			};
		},
	};
});

// Directive for checking if the user meets the necessary permissions
app.directive("arbUserCheck", function($compile, $mdToast, pageService, userService) {
	return {
		restrict: "A",
		controller: function($scope) {
			$scope.showUserCheckToast = function(message) {
				// TODO: restore when we figure out the bug with $mdToast
				//$mdToast.show($mdToast.simple().textContent(message));
			};
		},
		compile: function compile(element, attrs) {
			var check = attrs.arbUserCheck;
			var failMessage = "";
			if (!userService.user || userService.user.id === "0") {
				failMessage = "Login required";
			} else if (check === "cool") {
				if (!userService.userIsCool()) {
					failMessage = "You have a limited account";
				}
			}
			if (failMessage) {
				element.prepend(angular.element("<md-tooltip>" + failMessage + "</md-tooltip>"));
				attrs.ngClick = "showUserCheckToast('" + failMessage + "')";
			}
		},
	};
});

// Directive for a button to toggle requisite state
app.directive("arbRequisiteButton", function(pageService, userService) {
	return {
		templateUrl: "/static/html/requisiteButton.html",
		scope: {
			requisiteId: "@",
			// If true, don't show the page title
			hideTitle: "=",
			// If true, allow the user to toggle into a "want" state
			allowWants: "=",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function() {
				if (pageService.hasMastery($scope.requisiteId)) {
					if ($scope.allowWants) {
						pageService.updateMasteries([], [], [$scope.requisiteId]);
					} else {
						pageService.updateMasteries([], [$scope.requisiteId], []);
					}
				} else if (pageService.wantsMastery($scope.requisiteId)) {
					pageService.updateMasteries([], [$scope.requisiteId], []);
				} else {
					pageService.updateMasteries([$scope.requisiteId], [], []);
				}
			};
		},
	};
});
