'use strict';

// userName directive displayes a user's name.
app.directive('arbUserName', function(arb) {
	return {
		templateUrl: 'static/html/userName.html',
		scope: {
			userId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.user = arb.userService.userMap[$scope.userId];
		},
	};
});

// intrasitePopover contains the popover body html.
app.directive('arbIntrasitePopover', function($timeout, arb) {
	return {
		templateUrl: 'static/html/intrasitePopover.html',
		scope: {
			pageId: '@',
			direction: '@',
			arrowOffset: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
			$scope.summaries = [];
			$scope.getArrowStyle = function() {
				return {'left': +$scope.arrowOffset};
			};

			// We will check this to see if summaries are loaded.
			// Note that one-time binding takes effect after an object is set to something
			// other than undefined for the first time. So '::isLoaded' is safe, but '::!isLoaded'
			// is not safe (since it will be evaluated to true before isLoaded is set).
			$scope.isLoaded = undefined;
		},
		link: function(scope, element, attrs) {
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on('$destroy', function() {
				isDestroyed = true;
			});

			// Convert the name of the tab into an index for sorting.
			var nameToTabIndex = function(name) {
				if (name === 'Brief') return 0;
				if (name === 'Summary') return 1;
				if (name === 'Technical') return 2;
				return 3;
			};

			// Convert page's summaries into our local array
			var processPageSummaries = function() {
				if (!scope.page) return;
				for (var name in scope.page.summaries) {
					scope.summaries.push({name: name, text: scope.page.summaries[name]});
				}
				scope.summaries.sort(function(a, b) {
					return nameToTabIndex(a.name) > nameToTabIndex(b.name);
				});
				if (scope.summaries.length > 0) {
					scope.isLoaded = true;
				}
			};

			processPageSummaries();
			if (!scope.isLoaded) {
				// Fetch page summaries from the server.
				arb.pageService.loadIntrasitePopover(scope.pageId);
				// NOTE: we set up a watch instead of doing something on a success callback,
				// because the request might have been issued by another code already, and
				// in that case our callback wouldn't be called.
				var destroyWatcher = scope.$watch(function() {
					return scope.pageId in arb.stateService.pageMap ? Object.keys(arb.stateService.pageMap[scope.pageId].summaries).length : 0;
				}, function() {
					if (isDestroyed) {
						destroyWatcher();
						return;
					}
					scope.page = arb.stateService.pageMap[scope.pageId];
					processPageSummaries();
					if (scope.isLoaded) {
						destroyWatcher();
						// Hack: we need to fix the md-tabs height, because it takes way too long
						// to adjust by itself.
						$timeout(function() {
							var $el = element.find('.popover-tab-body');
							$el.closest('md-tabs').height($el.children().height());
						});
					}
				});
			}
		},
	};
});

// userPopover contains the popover body html.
app.directive('arbUserPopover', function($timeout, arb) {
	return {
		templateUrl: 'static/html/userPopover.html',
		scope: {
			userId: '@',
			direction: '@',
			arrowOffset: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.user = arb.userService.userMap[$scope.userId];
			$scope.page = arb.stateService.pageMap[$scope.userId];
			$scope.summaries = [];

			$scope.getArrowStyle = function() {
				return {'left': +$scope.arrowOffset};
			};

			// We will check this to see if summaries are loaded.
			// Note that one-time binding takes effect after an object is set to something
			// other than undefined for the first time. So '::isLoaded' is safe, but '::!isLoaded'
			// is not safe (since it will be evaluated to true before isLoaded is set).
			$scope.isLoaded = undefined;
		},
		link: function(scope, element, attrs) {
			// Fix to prevent errors when we go to another page while popover is loading.
			// TODO: abort all http requests when switching to another page
			var isDestroyed = false;
			scope.$on('$destroy', function() {
				isDestroyed = true;
			});

			// Convert page's summaries into our local array
			var processPageSummaries = function() {
				if (!scope.page || !scope.page.summaries) return;
				for (var name in scope.page.summaries) {
					scope.summaries.push({name: name, text: scope.page.summaries[name]});
				}
				if (scope.summaries.length > 0) {
					scope.isLoaded = true;
				}
			};

			processPageSummaries();
			if (!scope.isLoaded) {
				arb.userService.loadUserPopover(scope.userId);
				// NOTE: we set up a watch instead of doing something on a success callback,
				// because the request might have been issued by another code already, and
				// in that case our callback wouldn't be called.
				var destroyWatcher = scope.$watch(function() {
					return scope.userId in arb.stateService.pageMap ? Object.keys(arb.stateService.pageMap[scope.userId].summaries).length : 0;
				}, function() {
					if (isDestroyed) {
						destroyWatcher();
						return;
					}
					scope.user = arb.userService.userMap[scope.userId];
					scope.page = arb.stateService.pageMap[scope.userId];
					processPageSummaries();
					if (scope.isLoaded) {
						destroyWatcher();
						// Hack: we need to fix the md-tabs height, because it takes way too long
						// to adjust by itself.
						$timeout(function() {
							var $el = element.find('.popover-tab-body');
							$el.closest('md-tabs').height($el.children().height());
						});
					}
				});
			}
		},
	};
});

// pageTitle displays page's title with optional meta info.
app.directive('arbPageTitle', function(arb) {
	return {
		templateUrl: 'static/html/pageTitle.html',
		scope: {
			pageId: '@',
			// Options override for the page's title
			customPageTitle: '@',
			// Whether to display the title as a link or a span
			isLink: '=',
			// If set, we'll use this link for the page
			customLink: '@',
			// Whether or not to show the clickbait
			showClickbait: '=',
			// Whether or not to show the type of the page icon
			showType: '=',
			// If set, we'll pull the page from the edit map
			useEditMap: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.getPageFromSomeMap($scope.pageId, $scope.useEditMap);
			$scope.pageUrl = $scope.customLink ? $scope.customLink : arb.urlService.getPageUrl($scope.page.pageId);

			$scope.getTitle = function() {
				if ($scope.customPageTitle) {
					return $scope.customPageTitle;
				}
				return $scope.page.title;
			};
		},
	};
});

// likes displays the likes button(s) for a page.
app.directive('arbLikes', function($http, arb) {
	return {
		templateUrl: 'static/html/likes.html',
		scope: {
			// The type of likeable, such as 'changeLog'.
			likeableType: '@',
			// The id of the likeable object.
			likeableId: '@',
			// The likeable object this button corresponds to.
			// If likeableType is 'page', we'll look it up in the pageMap.
			likeable: '=',

			// If true, the button is not an icon button, but is a normal button
			isStretched: '=',
			// Whether or not we show likes as a button or a span
			isButton: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			if (!($scope.likeableType == 'page' || $scope.likeableType == 'changeLog')) {
				console.error('Unknown likeableType in arb-likes: ' + $scope.likeableType);
			}
			if (!$scope.likeable && $scope.likeableType == 'page') {
				$scope.likeable = arb.stateService.pageMap[$scope.likeableId];
			}

			// Sort individual likes by name.
			if ($scope.likeable && $scope.likeable.individualLikes) {
				$scope.likeable.individualLikes.sort(function(userId1, userId2) {
					return arb.userService.getFullName(userId1).localeCompare(arb.userService.getFullName(userId2));
				});
			}

			// User clicked on the like button
			$scope.likeClick = function() {
				if (!$scope.likeable) return;

				$scope.likeable.myLikeValue = Math.min(1, 1 - $scope.likeable.myLikeValue);

				var data = {
					likeableType: $scope.likeableType,
					id: $scope.likeableId,
					value: $scope.likeable.myLikeValue,
				};
				$http({method: 'POST', url: '/newLike/', data: JSON.stringify(data)})
				.error(function(data, status) {
					console.error('Error changing a like:'); console.log(data); console.log(status);
				});
			};
		},
	};
});

// subscribe directive displays the button for subscribing to a page.
app.directive('arbSubscribe', function($http, arb) {
	return {
		templateUrl: 'static/html/subscribe.html',
		scope: {
			pageId: '@',
			// If true, the button is not an icon button, but is a normal button with a label
			isStretched: '=',
			showSubscriberCount: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.isSubscribed = function() {
				return arb.stateService.pageMap[$scope.pageId].isSubscribed;
			};

			$scope.isSubscribedAsMaintainer = function() {
				return arb.stateService.pageMap[$scope.pageId].isSubscribedAsMaintainer;
			};

			// User clicked on the subscribe button
			$scope.subscribeClick = function() {
				arb.stateService.pageMap[$scope.pageId].isSubscribed = !$scope.isSubscribed();
				arb.stateService.pageMap[$scope.pageId].isSubscribedAsMaintainer = false;
				$http({
					method: 'POST',
					url: '/updateSubscription/',
					data: JSON.stringify({
						toId: $scope.pageId,
						isSubscribed: $scope.isSubscribed(),
						asMaintainer: false,
					})
				}).error(function(data, status) {
					console.error('Error changing a subscription:'); console.log(data); console.log(status);
				});
			};
		},
	};
});

// composeFab is the FAB button in the bottom right corner used for creating new pages
app.directive('arbComposeFab', function($location, $timeout, $mdMedia, $mdDialog, $rootScope, arb) {
	return {
		templateUrl: 'static/html/composeFab.html',
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.pageUrl = '/edit/';
			$scope.isSmallScreen = !$mdMedia('gt-sm');
			$scope.data = {
				isOpen: false,
			};
			$scope.showTooltips = arb.isTouchDevice;

			// Returns true if user has text selected on a touch device, and we should show
			// a special fab.
			$scope.showInlineVersion = function() {
				return arb.isTouchDevice && arb.stateService.lensTextSelected;
			};

			$scope.mouseEnter = function() {
				if (arb.isTouchDevice) return;
				$scope.data.isOpen = true;
			};

			$scope.mouseLeave = function() {
				if (arb.isTouchDevice) return;
				$scope.data.isOpen = false;
			};

			$scope.triggerClicked = function($event) {
				// Prevent angular material from doing its stuff.
				$event.stopPropagation();

				// If we're in the "inline response" mode, kick off the response.
				if ($scope.showInlineVersion()) {
					$rootScope.$broadcast('fabClicked');
					return;
				}

				// If it's open, execute the "New page" click.
				if ($scope.data.isOpen) {
					arb.urlService.goToUrl('/edit/');
				}

				// Toggle the menu.
				$scope.data.isOpen = !$scope.data.isOpen;
			};

			// Compute what the urls should be on the compose buttons, and which ones
			// should be visible.
			var computeUrls = function() {
				$scope.questionUrl = '/edit/?type=question';
				$scope.editPageUrl = undefined;
				$scope.childUrl = undefined;
				$scope.lensUrl = undefined;
				if (arb.stateService.primaryPage) {
					var type = arb.stateService.primaryPage.type;
					if (type === 'wiki' || type === 'group' || type === 'domain') {
						$scope.questionUrl = '/edit/?newParentId=' + arb.stateService.primaryPage.pageId + '&type=question';
						$scope.lensUrl = '/edit/?newParentId=' + arb.stateService.primaryPage.pageId + '&type=lens';
						$scope.childUrl = '/edit/?newParentId=' + arb.stateService.primaryPage.pageId;
					}
					if ($location.search().l) {
						$scope.editPageUrl = arb.urlService.getEditPageUrl($location.search().l);
					} else {
						$scope.editPageUrl = arb.urlService.getEditPageUrl(arb.stateService.primaryPage.pageId);
					}
				}
			};
			computeUrls();
			$scope.$watch(function() {
				// Note: can't use an object, so we just hack together a string
				return (arb.stateService.primaryPage ? arb.stateService.primaryPage.pageId : 'none') + $location.absUrl();
			}, function() {
				computeUrls();
			});

			// New feedback button is clicked
			$scope.newFeedback = function(event) {
				$mdDialog.show({
					templateUrl: 'static/html/feedbackDialog.html',
					controller: 'FeedbackDialogController',
					autoWrap: false,
					targetEvent: event,
				});
			};

			$scope.$on('$locationChangeSuccess', function() {
				$scope.hideFab = $location.path().indexOf('/edit') === 0;
			});
			$scope.hideFab = $location.path().indexOf('/edit') === 0;

			// Listen for shortcut keys
			$(document).keyup(function(event) {
				if (!event.ctrlKey || !event.altKey) return true;
				$scope.$apply(function() {
					if (event.keyCode == 80) arb.urlService.goToUrl('/edit/'); // P
					else if (event.keyCode == 69 && $scope.editPageUrl) arb.urlService.goToUrl($scope.editPageUrl); // E
					else if (event.keyCode == 67 && $scope.childUrl) arb.urlService.goToUrl($scope.childUrl); // C
					else if (event.keyCode == 78 && $scope.lensUrl) arb.urlService.goToUrl($scope.lensUrl); // N
					else if (event.keyCode == 81 && arb.stateService.primaryPage) $scope.newQueryMark(); // Q
					else if (event.keyCode == 75) $scope.newFeedback(event); // K
				});
			});

			$scope.newQueryMark = function() {
				$rootScope.$broadcast('newQueryMark');
			};
		},
	};
});

// autocomplete searches for relevant pages as you do the search
app.directive('arbAutocomplete', function($timeout, $q, arb) {
	return {
		templateUrl: 'static/html/autocomplete.html',
		scope: {
			// If true, the input will start out focused
			doAutofocus: '=',
			// Placeholder text
			placeholder: '@',
			// If set, the search will be constrained to this page type
			pageType: '@',
			// Function to call when a result is selected / user cancels selection
			onSelect: '&',
			// Function to call when input loses focus
			onBlur: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;

			// Called to get search results from the server
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				arb.autocompleteService.performSearch({term: text, pageType: $scope.pageType}, function(results) {
					deferred.resolve(results);
				});
				return deferred.promise;
			};

			// Called when user's choice changes
			$scope.ignoreNextResult = false;
			$scope.searchResultSelected = function(result) {
				var ignoring = $scope.ignoreNextResult;
				$scope.ignoreNextResult = false;
				if (ignoring) return;
				$scope.onSelect({result: result});
				// Note(alexei): this condition seems a little hacky, but it helps us prevent
				// calling callback twice.
				if ($scope.searchText || !result) {
					// Changing searchText will trigger this function, so we want to ignore it
					$scope.ignoreNextResult = true;
					$scope.searchText = '';
				}
			};
		},
		link: function(scope, element, attrs) {
			$timeout(function() {
				var $input = element.find("input");
				$input.on('blur', function(event) {
					if (scope.ignoreNextResult) return;
					// Make sure that if the user clicked one of the results, we don't count
					// it as a blur event.
					if (scope.onBlur && $(event.relatedTarget).closest('md-virtual-repeat-container').length <= 0) {
						scope.onBlur();
					}
				});
			});
		},
	};
});

// confirmButton is a button that ask for a confirmation when you press it
app.directive('arbConfirmButton', function(arb) {
	return {
		templateUrl: 'static/html/confirmButton.html',
		scope: {
			buttonText: '@',
			buttonBeforeConfirm: '@',
			disabled: '=',
			tooltipText: '@',
			confirmed: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.confirming = false;
			$scope.buttonFlexOrder = $scope.buttonBeforeConfirm ? -1 : 1;

			$scope.toggleConfirming = function(confirming) {
				$scope.confirming = confirming;
			};
		},
	};
});

// Directive for the User page panel
app.directive('arbPageList', function(arb) {
	return {
		templateUrl: 'static/html/pageList.html',
		scope: {
			pageIds: '=',
			panelTitle: '@',
			hideLikes: '=',
			showLastEdit: '=',
			showCreatedAt: '=',
			showQuickEdit: '=',
			showRedLinkCount: '=',
			showCommentCount: '=',
			showTextLength: '=',
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.getPage = function(pageId) {
				return arb.stateService.getPageFromSomeMap(pageId, $scope.useEditMap);
			};
		},
	};
});

// Exists to share the template for a row in a md-list of pages
app.directive('arbPageRow', function(arb) {
	return {
		templateUrl: 'static/html/pageRow.html',
		replace: true,
		scope: {
			pageId: '@',
			hideLikes: '=',
			showLastEdit: '=',
			showCreatedAt: '=',
			showQuickEdit: '=',
			showRedLinkCount: '=',
			showCommentCount: '=',
			showTextLength: '=',
			markAsDraft: '=',
			showTags: '=',
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.getPageFromSomeMap($scope.pageId, $scope.useEditMap);
		},
	};
});

app.directive('arbTag', function(arb) {
	return {
		template: '<span class="chip">{{tagName}}</span>',
		replace: true,
		scope: {
			tagId: '@',
		},
		controller: function($scope) {
			$scope.tagName = arb.stateService.getPageFromSomeMap($scope.tagId).title;
		},
	};
});

// Directive for checking if the user meets the necessary permissions
app.directive('arbUserCheck', function($compile, $mdToast, arb) {
	return {
		restrict: 'A',
		controller: function($scope) {
			$scope.showUserCheckToast = function(message) {
				arb.popupService.showToast({text: message, isError: true});
			};
		},
		compile: function compile(element, attrs) {
			var check = attrs.arbUserCheck;
			var failMessage = '';
			if (!arb.userService.userIsLoggedIn()) {
				failMessage = 'Login required';
			}
			if (failMessage) {
				element.prepend(angular.element('<md-tooltip md-direction="top">' + failMessage + '</md-tooltip>'));
				attrs.ngClick = 'showUserCheckToast(\'' + failMessage + '\')';
			}
		},
	};
});

// Directive for a button to toggle requisite state
app.directive('arbRequisiteButton', function(arb) {
	return {
		templateUrl: 'static/html/requisiteButton.html',
		scope: {
			requisiteId: '@',
			// If true, don't show the checkbox
			hideCheckbox: '=',
			// If true, don't show the page title
			hideTitle: '=',
			// If true, show requisite's clickbait
			showClickbait: '=',
			// If true, clicking the checkbox won't close the menu this button is in
			preventMenuClose: '=',
			// Optional callback function for when we change the mastery.
			unlockedFn: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;

			var unlockedCallback = undefined;
			if ($scope.unlockedFn) {
				unlockedCallback = function(data) {
					$scope.unlockedFn({result: data});
				};
			}

			// Toggle whether or not the user has a mastery
			$scope.toggleRequirement = function() {
				if (arb.masteryService.hasMastery($scope.requisiteId)) {
					arb.masteryService.updateMasteryMap({wants: [$scope.requisiteId], callback: unlockedCallback});
				} else if (arb.masteryService.wantsMastery($scope.requisiteId)) {
					arb.masteryService.updateMasteryMap({delete: [$scope.requisiteId], callback: unlockedCallback});
				} else {
					arb.masteryService.updateMasteryMap({knows: [$scope.requisiteId], callback: unlockedCallback});
				}
			};
		},
	};
});

// Directive for displaying next/prev buttons when learning.
app.directive('arbNextPrev', function($location, arb) {
	return {
		templateUrl: 'static/html/nextPrev.html',
		scope: {
			pageId: '@',
			// If true, show the expanded version of this directive
			extraInfo: '=',
			// If true, show the directive on a whiteframe
			whiteframe: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			$scope.stopLearning = function() {
				arb.pathService.abandonPath();
			};
		},
	};
});

// Directive for displaying individual changes in the changelog tab on the edit page.
app.directive('arbChangeLogEntry', function() {
	return {
		templateUrl: 'static/html/changeLogEntry.html',
	};
});

// Shared by the changelog and the updates page.
app.directive('arbLogRow', function(arb) {
	return {
		templateUrl: 'static/html/logRow.html',
		scope: {
			changeLog: '=', // Optional changelog associated with this row
			update: '=', // if this is shown in the updates page, this is the update object
			pageId: '@',
			byUserId: '@',
			type: '@',
			goToPageId: '@',
			isRelatedPageAlive: '=',
			markId: '@',
			createdAt: '@',
			repeated: '=',
			showUserLink: '=',
			showDismissIcon: '=',
			onDismiss: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.goToPage = $scope.goToPageId ? arb.stateService.pageMap[$scope.goToPageId] : undefined;
		},
	};
});
