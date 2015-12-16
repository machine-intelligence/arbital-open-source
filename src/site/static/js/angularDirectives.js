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
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.page = pageService.pageMap[$scope.pageId];

			// Add the primary page as the first lens.
			if ($scope.page && $scope.page.lensIds.indexOf($scope.page.pageId) < 0) {
				$scope.page.lensIds.unshift($scope.page.pageId);
			}

			// Check if a lens is loaded
			$scope.isLoaded = function(lensId) {
				var lens = pageService.pageMap[lensId];
				if (!lens) return false;
				return lens.summary.length > 0;
			};
		},
		link: function(scope, element, attrs) {
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on("$destroy", function() {
				isDestroyed = true;
			});

			// Called when a tab is selected
			scope.tabSelect = function(lensId) {
				if (scope.isLoaded(lensId)) return;
				// Fetch page data from the server.
				pageService.loadIntrasitePopover(lensId, {
					success: function() {
						if (isDestroyed) return;
						if (!scope.page) {
							scope.page = pageService.pageMap[scope.pageId];
						}
						var lens = pageService.pageMap[lensId];
						if (!lens.summary) {
							lens.summary = " "; // to avoid trying to load it again
						}
						// Hack: we need to fix the md-tabs height, because it takes way too long
						// to adjust by itself.
						$timeout(function() {
							var $el = element.find(".popover-tab-body");
							$el.closest("md-tabs").height($el.children().height());
						});
						// Page's lensIds got resent, so need to fix this again
						if (scope.page.lensIds.indexOf(scope.page.pageId) < 0) {
							scope.page.lensIds.unshift(scope.page.pageId);
						}
					},
				});
			};
			scope.tabSelect(scope.pageId);
		},
	};
});

// userPopover contains the popover body html.
app.directive("arbUserPopover", function(pageService, userService) {
	return {
		templateUrl: "/static/html/userPopover.html",
		scope: {
			userId: "@",
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;
			$scope.user = userService.userMap[$scope.userId];
			
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			$scope.$on("$destroy", function() {
				isDestroyed = true;
			});

			// Check if the data is loaded
			$scope.isLoaded = function() {
				if (!($scope.userId in userService.userMap)) {
					return false;
				}
				if (!($scope.userId in pageService.pageMap)) {
					return false;
				}
				return pageService.pageMap[$scope.userId].summary.length > 0;
			};

			if (!$scope.isLoaded()) {
				pageService.loadUserPopover($scope.userId, {
					success: function() {
						var userPage = pageService.pageMap[$scope.userId];
						if (userPage && userPage.summary.length == 0) {
							userPage.summary = " "; // to prevent from loading again
						}
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
		},
		controller: function($scope) {
			$scope.pageService = pageService;
			$scope.userService = userService;

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
app.directive("arbComposeFab", function($location, $timeout, $mdMedia, pageService, userService) {
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
				$scope.siblingUrl = undefined;
				$scope.childUrl = undefined;
				$scope.lensUrl = undefined;
				$scope.showNewComment = false;
				$scope.showNewAnswer = false;
				if (pageService.primaryPage) {
					$scope.editPageUrl = pageService.getEditPageUrl(pageService.primaryPage.pageId);
					var type = pageService.primaryPage.type;
					if (type === "question") {
						$scope.showNewAnswer = true;
					} else if (type === "wiki") {
						$scope.showNewComment = true;
						$scope.questionUrl = "/edit/?newParentId=" + pageService.primaryPage.pageId + "&type=question";
						$scope.lensUrl = "/edit/?newParentId=" + pageService.primaryPage.pageId + "&type=lens";
						$scope.childUrl = "/edit?newParentId=" + pageService.primaryPage.pageId;
						if (pageService.primaryPage.parentIds.length > 0) {
							$scope.siblingUrl = "/edit?newParentId=" + pageService.primaryPage.parentIds.join(",");
						}
					}
				}
			};
			computeUrls();
			$scope.$watch(function() { return pageService.primaryPage; }, function() {
				computeUrls();
			});

			// New comment button is clicked
			$scope.newComment = function() {
				$("html, body").animate({
					scrollTop: $(".new-comment-button").offset().top,
		    }, {
					duration: 1000,
					complete: function() {
						$(".new-comment-button").click();
					}
				});
			};

			// New answer button is clicked
			$scope.newAnswer = function() {
				$("html, body").animate({
					scrollTop: $("#your-answer").offset().top,
		    }, 1000);
			};

			$scope.$on("$locationChangeSuccess", function () {
				$scope.hide = $location.path().indexOf("/edit") === 0;
			});
			$scope.hide = $location.path().indexOf("/edit") === 0;
		},
	};
});

// autocomplete searches for relevant pages as you do the search
app.directive("arbAutocomplete", function($timeout, $q, pageService, autocompleteService) {
	return {
		templateUrl: "/static/html/autocomplete.html",
		scope: {
			doAutofocus: "=",
			placeholder: "@",
			// If set, the search will be constrained to this page type
			pageType: "@",
			onSelect: "&",
			onBlur: "&",
		},
		controller: function($scope) {
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
				$scope.searchText = "";
			};
		},
		link: function(scope, element, attrs) {
			$timeout(function() {
				$(element).find("input").blur(function() {
					if (scope.onBlur) scope.onBlur();
				});
			});
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
			// Whether to show a public/private icon, pass in "yes"/"no" respectively.
			isPublic: "@",
			hideLikes: "@",
			showLastEdit: "@",
			showCreatedAt: "@",
			showQuickEdit: "@",
			showRedLinkCount: "@",
			showCommentCount: "@",
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: "@",
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
				$mdToast.show($mdToast.simple().textContent(message));
			};
		},
		compile: function compile(element, attrs) {
			var check = attrs.arbUserCheck;
			var failMessage = "";
			if (userService.user.id === "0") {
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
