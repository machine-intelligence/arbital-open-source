'use strict';

import app from './angular.ts';

// userName directive displayes a user's name.
app.directive('arbUserName', function(arb) {
	return {
		templateUrl: versionUrl('static/html/userName.html'),
		scope: {
			userId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.user = arb.userService.userMap[$scope.userId];
		},
	};
});

// arb-click-elsewhere calls a function when there is a click outside the element
app.directive('arbClickElsewhere', function ($parse, $window) {
	return {
		restrict: 'A',
		link: function (scope, element, attr) {
			var element = element;
			var onClickElsewhereFn = $parse(attr['arbClickElsewhere']);
			angular.element($window).on('click', function (event) {
				var clickTarget = angular.element(event.target);
				if (!element.is(clickTarget) && !element[0].contains(clickTarget[0])) {
					onClickElsewhereFn(scope);
				}
			});
		}
	};
});

// arb-edit-button shows an edit button for a page, and handles users not being logged in
app.directive('arbEditButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/editButton.html'),
		scope: {
			pageId: '@',
			createText: '=',
			analyticsDesc: '@',
			customText: '@',
			hideIcon: '=',
			primary: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.processClick = function(event) {
				arb.analyticsService.reportEditPageAction(event, $scope.analyticsDesc);
				arb.signupService.wrapInSignupFlow('edit click:' + $scope.analyticsDesc,
					function() {
						arb.urlService.goToUrl(arb.urlService.getEditPageUrl($scope.pageId), {event: event});
					});
			};

			$scope.getButtonText = function() {
				if ($scope.customText !== undefined) return $scope.customText;
				if (!arb.userService.userIsLoggedIn() || !$scope.page.permissions.edit.has) return 'Propose edit';
				if ($scope.page.hasDraft) return 'Edit draft';
				if ($scope.page.proposalEditNum) return 'Review proposal';
				return 'Edit';
			};
		},
	};
});

// directive for an expanded icon
app.directive('arbExpandIcon', function(arb) {
	return {
		templateUrl: versionUrl('static/html/expandIcon.html'),
		scope: {
			expanded: '=',
		},
	};
});

// directive for a sub-header in a list
app.directive('arbListSubHeader', function(arb) {
	return {
		templateUrl: versionUrl('static/html/listSubHeader.html'),
		transclude: true,
	};
});

// intrasitePopover contains the popover body html.
app.directive('arbIntrasitePopover', function($timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/intrasitePopover.html'),
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

			$scope.onSwipe = function() {
				if (!arb.isTouchDevice) return;
				arb.popoverService.removePopover();
			};
		},
		link: function(scope: any, element, attrs) {
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
		templateUrl: versionUrl('static/html/userPopover.html'),
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

			$scope.onSwipe = function() {
				if (!arb.isTouchDevice) return;
				arb.popoverService.removePopover();
			};
		},
		link: function(scope: any, element, attrs) {
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

// textPopover contains the popover body html.
app.directive('arbTextPopover', function($compile, $timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/textPopover.html'),
		scope: {
			encodedHtml: '@',
			direction: '@',
			arrowOffset: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.getArrowStyle = function() {
				return {'left': +$scope.arrowOffset};
			};

			$scope.onSwipe = function() {
				if (!arb.isTouchDevice) return;
				arb.popoverService.removePopover();
			};
		},
		link: function(scope: any, element, attrs) {
			element.find('.popover-tab-body').html(decodeURIComponent(scope.encodedHtml));
			arb.markdownService.compileChildren(scope, element);
		},
	};
});

// arb-text-popover-anchor is the thing you can hover over to get a text popover
app.directive('arbTextPopoverAnchor', function($timeout, arb) {
	return {
		scope: {
			popoverText: '@',
			replacementText: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
		},
		link: function(scope: any, element, attrs) {
			if (!scope.popoverText) {
				element.attr('encoded-html', encodeURIComponent(element.html()));
				element.text(scope.replacementText || '!');
			} else {
				element.attr('encoded-html', encodeURIComponent(scope.popoverText));
			}
		},
	};
});

// RedLinkPopover is shown when the user hovers over a red link
app.directive('arbRedLinkPopover', function($timeout, arb) {
	return {
		templateUrl: versionUrl('static/html/redLinkPopover.html'),
		scope: {
			alias: '@',
			isRedText: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.redLinkRow = undefined;

			$scope.onSwipe = function() {
				if (!arb.isTouchDevice) return;
				arb.popoverService.removePopover();
			};

			// Fetch data from the server.
			if (!$scope.isRedText) {
				arb.pageService.loadRedLinkPopover($scope.alias, {
					success: function(data) {
						$scope.redLinkRow = data.result.redLinkRow;
					},
				});
			}

			// Get how total number of views between all the pages that use this red link.
			$scope.getRedLinkViews = function() {
				let totalViews = 0;
				for (let pageId of $scope.redLinkRow.linkedByPageIds) {
					totalViews += arb.stateService.pageMap[pageId].viewCount;
				}
				return totalViews;
			};
		},
	};
});

// pageTitle displays page's title with optional meta info.
app.directive('arbPageTitle', function(arb) {
	return {
		templateUrl: versionUrl('static/html/pageTitle.html'),
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
			// The class to use for the clickbait (if not the default)
			clickbaitClass: '@',
			// Whether or not to show the type of the page icon
			showType: '=',
			// If true, show information about this being an arc
			showArcInfo: '=',
			// If set, we'll pull the page from the edit map
			useEditMap: '=',
			// If true, link to editing the page
			linkToEdit: '=',
			// If set, the link will be computed with this hubId
			hubId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.pageUrl = $scope.customLink ? $scope.customLink :
							 $scope.linkToEdit ? arb.urlService.getEditPageUrl($scope.pageId) :
							 		arb.urlService.getPageUrl($scope.pageId, {hubId: $scope.hubId});

			$scope.page = arb.stateService.getPageFromSomeMap($scope.pageId, $scope.useEditMap);

			$scope.getTitle = function() {
				let title = $scope.customPageTitle || $scope.page.title;
				if ($scope.showArcInfo && $scope.page.pathPages.length > 0) {
					title += ' (' + $scope.page.pathPages.length + ' page arc)';
				}
				return title;
			};
		},
	};
});

// likes displays the likes button(s) for a page.
app.directive('arbLikes', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/likes.html'),
		scope: {
			// ObjectId of the likeable object.
			objectId: '@',
			// The likeable object this button corresponds to. If it's not set, we assume
			// objectId is a pageId and fetch the corresponding page.
			likeable: '=',
			// If true, the button is not an icon button, but is a normal button
			isStretched: '=',
			// Whether or not we show likes as a button or a span
			isSpan: '=',
			// If true, show the +1 version of the button
			isPlusOne: '=',
			// Optional function to call when the user clicks the like button.
			onLikeClicked: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;

			if (!$scope.likeable) {
				$scope.likeable = arb.stateService.pageMap[$scope.objectId];
			}

			// Sort individual likes by name.
			// TODO: move this to BE, otherwise we are sorting this array each time an
			// instance of the like button is created.
			if ($scope.likeable && $scope.likeable.individualLikes) {
				$scope.likeable.individualLikes.sort(function(userId1, userId2) {
					return arb.userService.getFullName(userId1).localeCompare(arb.userService.getFullName(userId2));
				});
			}

			// Called when the like button is clicked
			$scope.likeClicked = function() {
				arb.signupService.processLikeClick($scope.likeable, $scope.objectId);
				if ($scope.onLikeClicked) {
					$scope.onLikeClicked({result: {
						myLikeValue: $scope.likeable.myLikeValue,
					}});
				}
			};
		},
	};
});

// subscribe directive displays the button for subscribing to a page.
app.directive('arbSubscribe', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/subscribe.html'),
		scope: {
			pageId: '@',
			// If true, the button is not an icon button, but is a normal button with a label
			isStretched: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.isSubscribed = function() {
				return arb.stateService.pageMap[$scope.pageId].isSubscribed;
			};

			// User clicked on the subscribe button
			$scope.subscribeClick = function() {
				arb.stateService.pageMap[$scope.pageId].isSubscribed = !$scope.isSubscribed();
				let postParams = {
					toId: $scope.pageId,
					isSubscribed: $scope.isSubscribed(),
				};
				arb.stateService.postDataWithoutProcessing('/updateSubscription/', postParams);
			};
		},
	};
});

// subscribe directive displays the button for subscribing to a page as a maintainer.
app.directive('arbSubscribeToMaintain', function($http, arb) {
	return {
		templateUrl: versionUrl('static/html/subscribeToMaintain.html'),
		scope: {
			pageId: '@',
			// If true, the button is not an icon button, but is a normal button with a label
			isStretched: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];

			$scope.isSubscribed = function() {
				return arb.stateService.pageMap[$scope.pageId].isSubscribedAsMaintainer;
			};

			// User clicked on the subscribe button
			$scope.subscribeClick = function() {
				arb.stateService.pageMap[$scope.pageId].isSubscribedAsMaintainer  = !$scope.isSubscribed();
				let postParams = {
					toId: $scope.pageId,
					isSubscribed: $scope.isSubscribed(),
					asMaintainer: true,
				};
				arb.stateService.postDataWithoutProcessing('/updateSubscription/', postParams);
			};
		},
	};
});

// composeFab is the FAB button in the bottom right corner used for creating new pages
app.directive('arbComposeFab', function($location, $timeout, $mdMedia, $mdDialog, $rootScope, arb) {
	return {
		templateUrl: versionUrl('static/html/composeFab.html'),
		scope: {
		},
		controller: function($scope) {
			$scope.arb = arb;
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

			$scope.triggerClicked = function(event) {
				// Prevent angular material from doing its stuff.
				event.stopPropagation();

				// If we're in the "inline response" mode, kick off the response.
				if ($scope.showInlineVersion()) {
					$rootScope.$broadcast('fabClicked');
					return;
				}

				// If it's open, execute the "New page" click.
				if ($scope.data.isOpen) {
					arb.urlService.goToUrl($scope.newPageUrl, {event: event});
				}

				// Toggle the menu.
				$scope.data.isOpen = !$scope.data.isOpen;
			};

			// Compute what the urls should be on the compose buttons, and which ones
			// should be visible.
			var computeUrls = function() {
				$scope.newPageUrl = arb.urlService.getNewPageUrl({
					parentId: arb.stateService.primaryPage ? arb.stateService.primaryPage.pageId : undefined,
				});

				$scope.editPageUrl = undefined;
				if (arb.stateService.primaryPage) {
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
					templateUrl: versionUrl('static/html/feedbackDialog.html'),
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
					if (event.keyCode == 80) arb.urlService.goToUrl($scope.newPageUrl); // P
					else if (event.keyCode == 69 && $scope.editPageUrl) arb.urlService.goToUrl($scope.editPageUrl); // E
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
		templateUrl: versionUrl('static/html/autocomplete.html'),
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
			// If true, only search over groups
			searchGroups: '=',
			// If true, exclude groups from search results
			ignoreGroups: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;

			// Called to get search results from the server
			$scope.getSearchResults = function(text) {
				if (!text) return [];
				var deferred = $q.defer();
				if ($scope.searchGroups) {
					arb.autocompleteService.userSource({term: text}, function(results) {
						deferred.resolve(results);
					});
				} else {
					arb.autocompleteService.performSearch({
						term: text,
						pageType: $scope.pageType,
						filterPageTypes: $scope.ignoreGroups ? ['group'] : [],
					}, function(results) {
						deferred.resolve(results);
					});
				}
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

			$scope.searchTextToAlias = function(searchText) {
				return searchText.split(' ').join('_');
			};
		},
		link: function(scope: any, element, attrs) {
			$timeout(function() {
				var $input = element.find('input');
				$input.on('blur', function(event) {
					if (scope.ignoreNextResult) return;
					// Make sure that if the user clicked one of the results, we don't count
					// it as a blur event.
					// NOTE: that Firefox doesn't have event.relatedTarget
					var relatedTarget = event.relatedTarget || event.originalEvent.explicitOriginalTarget;
					if (scope.onBlur && $(relatedTarget).closest('md-virtual-repeat-container').length <= 0) {
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
		templateUrl: versionUrl('static/html/confirmButton.html'),
		scope: {
			buttonText: '@',
			buttonBeforeConfirm: '@',
			disabled: '=',
			tooltipText: '@',
			confirmed: '&',
			checkerFn: '&',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.confirming = false;
			$scope.buttonFlexOrder = $scope.buttonBeforeConfirm ? -1 : 1;

			$scope.toggleConfirming = function(confirming) {
				if (!$scope.confirming) {
					// Check if we need to confirm
					if ($scope.checkerFn) {
						if (!$scope.checkerFn()) {
							$scope.confirmed();
						}
					}
				}
				$scope.confirming = confirming;
			};
		},
	};
});

// Directive for the User page panel
app.directive('arbPageList', function(arb) {
	return {
		templateUrl: versionUrl('static/html/pageList.html'),
		scope: {
			pageIds: '=',
			panelTitle: '@',
			// How many more items to load if more items are requested for the list
			loadItemsIncrement: '=',
			hideLikes: '=',
			showLastEdit: '=',
			showCreatedAt: '=',
			showQuickEdit: '=',
			showRedLinkCount: '=',
			showCommentCount: '=',
			showTextLength: '=',
			// If set, we'll pull the page from the editMap instead of pageMap
			useEditMap: '=',
			// If set, we'll call this end-point to fetch the data
			sourceUrl: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.loadItemsTotal = $scope.loadItemsIncrement;

			$scope.getPage = function(pageId) {
				return arb.stateService.getPageFromSomeMap(pageId, $scope.useEditMap);
			};

			$scope.fetchMore = function() {
				$scope.loadItemsTotal += $scope.loadItemsIncrement;
				$scope.fetchingMore = true;
				arb.stateService.postData($scope.sourceUrl, {
						numToLoad: $scope.loadItemsTotal
					},
					function(data) {
						$scope.pageIds = data.result.pageIds;
						$scope.fetchingMore = false;
					}
				);
			};

			if ($scope.sourceUrl && (!$scope.pageIds || $scope.pageIds.length <= 0)) {
				$scope.fetchMore();
			}
		},
	};
});

// Exists to share the template for a row in a md-list of pages
app.directive('arbPageRow', function($mdMedia, arb) {
	return {
		templateUrl: versionUrl('static/html/rows/pageRow.html'),
		replace: true,
		scope: {
			pageId: '@',
			hideLikes: '=',
			showLastEdit: '=',
			showCreatedAt: '=',
			showCreatedBy: '=',
			showOtherDateTime: '=',
			otherDateTime: '=',
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

			$scope.isTinyScreen = !$mdMedia('gt-xs');
		},
	};
});

app.directive('arbTag', function(arb) {
	return {
		template: '<a ng-href="{{url}}" class="chip">{{tagName}}</a>',
		replace: true,
		scope: {
			tagId: '@',
		},
		controller: function($scope) {
			var tag = arb.stateService.getPage($scope.tagId);
			$scope.tagName = tag ? tag.title : $scope.tagId;
			$scope.url = arb.urlService.getPageUrl($scope.tagId);
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
		compile: function compile(element, attrs: any) {
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
// OBSOLETE
/*
app.directive('arbRequisiteButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/requisiteButton.html'),
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
});*/

// Directive for displaying next/prev buttons when learning.
app.directive('arbNextPrev', function($location, arb) {
	return {
		templateUrl: versionUrl('static/html/nextPrev.html'),
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
		templateUrl: versionUrl('static/html/changeLogEntry.html'),
	};
});

// Row to show in the changelog tab
app.directive('arbChangeLogRow', function(arb) {
	return {
		templateUrl: versionUrl('static/html/rows/changeLogRow.html'),
		scope: {
			changeLog: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.byUserId = $scope.changeLog.userId;
			$scope.auxPageId = $scope.changeLog.auxPageId;
		},
	};
});

app.directive('arbLensToolbar', function($window, $mdConstant, $mdUtil, $compile, $timeout, $http, arb) {
	return {
		templateUrl: versionUrl('static/html/lensToolbarWrapper.html'),
		scope: false,
		controller: function($scope) {
			$scope.arb = arb;
			$scope.noFloater = true;

			// Process click on "Subscribe as maintainer"
			$scope.toggleMaintainerSub = function() {
				$scope.page.isSubscribedAsMaintainer = !$scope.page.isSubscribedAsMaintainer;
				if ($scope.page.isSubscribedAsMaintainer) {
					$scope.page.isSubscribed = true;
				}

				$http({method: 'POST', url: '/updateSubscription/', data: JSON.stringify({
					toId: $scope.page.pageId,
					isSubscribed: $scope.page.isSubscribed,
					asMaintainer: $scope.page.isSubscribedAsMaintainer,
				})});
			};
		},
		link: function(scope: any, element) {
			var staticBar = angular.element(element.find('#static-toolbar'));
			var floaterBar = angular.element(element.find('#floater-toolbar'));

			const scrollUpBuffer = arb.isTouchDevice ? 50 : 10;

			// Control the width of the floater bar
			var setFloaterWidth = function() {
				floaterBar.css('width', staticBar.css('width'));
			};
			angular.element($window).bind('resize', setFloaterWidth);
			staticBar.bind('resize', setFloaterWidth);

			// Control the behavior of the floater bar
			var prevWindowY; // Used to tell if we're scrolling up or down.
			var yFromFirstScrollUp; // The y coordinate from when we first started scrolling up.
			var scrollingDown; // Whether we are scrolling up or down.
			var onScroll = function() {
				// Figure out whether we're scrolling down
				var wasScrollingDown = scrollingDown;
				var currWindowY = $window.scrollY;
				scrollingDown = currWindowY > prevWindowY;
				prevWindowY = currWindowY;

				// If we were scrolling down and we are not anymore, record the y position
				if (wasScrollingDown && !scrollingDown) {
					yFromFirstScrollUp = currWindowY;
				}

				// If we're scrolling down or we have only scrolled up a little, hide the floater
				scope.hideFloater = scrollingDown || (yFromFirstScrollUp - currWindowY  < scrollUpBuffer);

				// If the bottom of the staticBar is visible, hide the floater bar completely
				var staticBarVisible = staticBar[0].getBoundingClientRect().bottom <=
						document.documentElement.clientHeight;
				scope.noFloater = staticBarVisible;
			};
			angular.element($window).bind('scroll', onScroll);

			var setUpLensToolbar = function() {
				$timeout(function() {
					prevWindowY = 1000;
					onScroll();
					setFloaterWidth();
				});
			};
			setUpLensToolbar();
		},
	};
});

app.directive('arbSlackButton', function(arb) {
	return {
		templateUrl: versionUrl('static/html/slackButton.html'),
		scope: {
			// Button text when the user hasn't joined slack yet
			joinText: '@',
			// Button text when the user already joined slack
			memberText: '@',
			raised: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			if (!$scope.joinText) {
				$scope.joinText = 'Join community chat';
			}
			if (!$scope.memberText) {
				$scope.memberText = 'Go to Slack';
			}
			$scope.isUserMember = (arb.userService.user && arb.userService.user.isSlackMember) ||
				Cookies.getJSON('isSlackMember');

			// Called when user clicks join slack button
			$scope.joinSlack = function() {
				if ($scope.isUserMember) {
					arb.analyticsService.reportEventToHeapAndMixpanel('goto slack', {
						pageId: arb.stateService.primaryPage ? arb.stateService.primaryPage.pageId : undefined,
					});
					window.open('https://arbital.slack.com/messages/math/', '_blank');
				} else {
					arb.analyticsService.reportEventToHeapAndMixpanel('join slack', {
						pageId: arb.stateService.primaryPage ? arb.stateService.primaryPage.pageId : undefined,
					});
					arb.signupService.wrapInSignupFlow('join slack', function() {
						var postParams = {email: arb.userService.user.email};
						arb.stateService.postDataWithoutProcessing('/json/sendSlackInvite/', postParams, function() {
							$scope.isUserMember = true;
							arb.userService.user.isSlackMember = true;
							Cookies.set('isSlackMember', true);
							arb.popupService.showToast({text: 'Thanks! You should get an invite within a few hours.'});
						});
					});
				}
			};
		},
	};
});

// Dropdown for selecting various domain roles.
app.directive('arbDomainRoleInput', function(arb) {
	return {
		templateUrl: versionUrl('static/html/domainRoleInput.html'),
		scope: {
			domainId: '@',
			// Option 1. If given, this user's role will be shown and changed.
			userId: '@',
			// Option 2. If given, the role input will be stored here.
			domainMembership: '=',
			showLabel: '=',
		},
		controller: function($scope) {
			$scope.arb = arb;
			if ($scope.userId) {
				$scope.domainMembership = arb.userService.userMap[$scope.userId].domainMembershipMap[$scope.domainId];
			}

			$scope.updateDomainRole = function() {
				if (!$scope.userId) return;
				let data = {
					userId: $scope.userId,
					domainId: $scope.domainId,
					role: $scope.domainMembership.role,
				};
				arb.stateService.postDataWithoutProcessing('/updateDomainRole/', data);
			};
		},
	};
});

// Showing a link to comment section along with comment count
app.directive('arbCommentCount', function(arb) {
	return {
		templateUrl: versionUrl('static/html/commentCount.html'),
		scope: {
			pageId: '@',
		},
		controller: function($scope) {
			$scope.arb = arb;
			$scope.page = arb.stateService.pageMap[$scope.pageId];
		},
	};
});
