'use strict';

// ArbitalCtrl is the top level controller. It never gets unloaded.
app.controller('ArbitalCtrl', function($rootScope, $scope, $location, $timeout, $interval, $http, $compile, $anchorScroll, $mdDialog, arb) {
	$scope.arb = arb;

	// Refresh all the fields that need to be updated every so often.
	var refreshAutoupdates = function() {
		$('.autoupdate').each(function(index, element) {
			$compile($(element))($scope);
		});
		$timeout(refreshAutoupdates, 30000);
	};
	refreshAutoupdates();

	// Don't allow the body to scroll when scrolling a popover tab body
	// http://stackoverflow.com/questions/5802467/prevent-scrolling-of-parent-element
	if (!arb.isTouchDevice) {
		$(document).on('DOMMouseScroll mousewheel', '.prevent-scroll-leak', function(ev) {
			var $this = $(this);
			var scrollTop = this.scrollTop;
			var scrollHeight = this.scrollHeight;
			var height = $this.innerHeight();
			var delta = (ev.type == 'DOMMouseScroll' ?	ev.originalEvent.detail * -40 : ev.originalEvent.wheelDelta);
			var up = delta > 0;
			// Don't prevent body scrolling if there is no scroll bar
			if (scrollHeight <= this.clientHeight) return true;

			var prevent = function() {
				ev.stopPropagation();
				ev.preventDefault();
				ev.returnValue = false;
				return false;
			};

			if (!up && -delta > scrollHeight - height - scrollTop) {
				// Scrolling down, but this will take us past the bottom.
				$this.scrollTop(scrollHeight);
				return prevent();
			} else if (up && delta > scrollTop) {
				// Scrolling up, but this will take us past the top.
				$this.scrollTop(0);
				return prevent();
			}
		});
	}

	// Check to see if we should show the popup.
	$scope.closePopup = function() {
		arb.popupService.hideNonpersistentPopup();
	};

	// Watch path changes and update Google Analytics
	$scope.$watch(function() {
		return $location.absUrl();
	}, function() {
		ga('send', 'pageview', $location.absUrl());
	});

	var $fixedOverlay = $('#fixed-overlay');
	$scope.$watch(function() {
		return $fixedOverlay.children().length;
	}, function() {
		// If we have md-scroll-mask on, we want the fixedOverlay to occupy entire screen
		// so that the mdBottomSheet UI is displayed fully.
		var hasScrollMask = $fixedOverlay.find('.md-scroll-mask').length > 0;
		$fixedOverlay.css('height', hasScrollMask ? '100vh' : '0');
	});

	// Returns an object containing a compiled element and its scope
	$scope.newElement = function(html, parentScope) {
		if (!parentScope) parentScope = $scope;
		var childScope = parentScope.$new();
		var element = $compile(html)(childScope);
		return {
			scope: childScope,
			element: element,
		};
	};
	// The element and it scope inside ng-view for the current page
	var currentView;

	// Returns a function we can use as success handler for POST requests for dynamic data.
	// callback - returns {
	//	 title: title to set for the window
	//	 element: optional jQuery element to add dynamically to the body
	//	 error: optional error message to print
	// }
	$scope.getSuccessFunc = function(callback) {
		return function(data) {
			// Sometimes we don't get data.
			arb.stateService.primaryPage = undefined;
			if (data) {
				console.log('Dynamic request data:'); console.log(data);
				arb.stateService.processServerData(data);
			}

			// Because the subdomain could have any case, we need to find the alias
			// in the loaded map so we can get the alias with correct case
			if ($scope.subdomain) {
				for (var pageAlias in arb.stateService.pageMap) {
					if ($scope.subdomain.toUpperCase() === pageAlias.toUpperCase()) {
						$scope.subdomain = pageAlias;
						arb.stateService.privateGroupId = arb.stateService.pageMap[pageAlias].pageId;
						break;
					}
				}
			}

			if (currentView) {
				currentView.scope.$destroy();
				currentView.element.remove();
				currentView = null;
				arb.urlService.hasLoadedFirstPage = true;
			}

			// Get the results from page-specific callback
			$('.global-error').hide();
			var result = callback(data);
			if (result.error) {
				$('.global-error').text(result.error).show();
				document.title = 'Error - Arbital';
			}

			if (result.content) {
				// Only show the element after it and all the children have been fully compiled and linked
				result.content.element.addClass('reveal-after-render-parent');
				var $loadingBar = $('#loading-bar');
				$loadingBar.show();
				var startTime = (new Date()).getTime();

				var showEverything = function() {
					$interval.cancel(revealInterval);
					$timeout.cancel(revealTimeout);
					// Do short timeout to prevent some rendering bugs that occur on edit page
					$timeout(function() {
						result.content.element.removeClass('reveal-after-render-parent');
						$loadingBar.hide();
					}, 50);
				};

				var revealInterval = $interval(function() {
					var timePassed = ((new Date()).getTime() - startTime) / 1000;
					var hiddenChildren = result.content.element.find('.reveal-after-render');
					if (hiddenChildren.length > 0) {
						hiddenChildren.each(function() {
							if ($(this).children().length > 0) {
								$(this).removeClass('reveal-after-render');
							}
						});
						return;
					}
					showEverything();
				}, 50);
				// Do a timeout as well, just in case we have a buggy element
				var revealTimeout = $timeout(function() {
					console.error('Forced reveal timeout');
					showEverything();
				}, 1000);

				currentView = result.content;
				$('[ng-view]').append(result.content.element);
			}

			$('body').toggleClass('body-fix', !result.removeBodyFix);

			if (result.title) {
				document.title = result.title + ' - Arbital';
			}
		};
	};

	// Returns a function we can use as error handler for POST requests for dynamic data.
	$scope.getErrorFunc = function(urlPageType) {
		return function(data, status) {
			console.error('Error /json/' + urlPageType + '/:'); console.log(data); console.log(status);
			arb.popupService.showToast({text: 'Couldn\'t create a new page: ' + data, isError: true});
			document.title = 'Error - Arbital';
		};
	};

	// The URL rule match for the current page
	var currentLocation = {};
	function resolveUrl() {
		// Get subdomain if any
		$scope.subdomain = undefined;
		var subdomainMatch = /^([A-Za-z0-9_]+)\.(localhost|arbital\.com)\/?$/.exec($location.host());
		if (subdomainMatch) {
			$scope.subdomain = subdomainMatch[1];
		}
		var path = $location.path();
		var urlRules = arb.urlService.urlRules;
		for (var ruleIndex = 0; ruleIndex < urlRules.length; ruleIndex++) {
			var rule = urlRules[ruleIndex];
			var matches = rule.urlPattern.exec(path);
			if (matches) {
				var args = {};
				var parameters = rule.parameters;
				for (var parameterIndex = 0; parameterIndex < parameters.length; parameterIndex++) {
					var parameter = parameters[parameterIndex];
					args[parameter] = matches[parameterIndex + 1];
				}
				if (rule == currentLocation.rule && $scope.subdomain == currentLocation.subdomain) {
					var currentMatches = true;
					for (parameterIndex = 0; parameterIndex < parameters.length && currentMatches; parameterIndex++) {
						var parameter = parameters[parameterIndex];
						currentMatches = (args[parameter] == currentLocation.args[parameter]);
					}
					if (currentMatches) {
						// The host and path have not changed, don't reload
						return;
					}
				}
				var handled = rule.handler(args, $scope);
				if (!handled) {
					$('[ng-view]').empty();
					$scope.closePopup();
				}
				currentLocation = {subdomain: $scope.subdomain, rule: rule, args: args};
				return;
			}
		}
	};

	$rootScope.$on('$locationChangeSuccess', function(event, url) {
		resolveUrl();
	});

	// Resolve URL of initial page load
	resolveUrl();
});

