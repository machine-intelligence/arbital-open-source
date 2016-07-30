'use strict';

import app from './angular.ts';
import {isTouchDevice} from './util.ts';

// Popover service is used to display the intrasite popover.
app.service('popoverService', function($rootScope, $compile, $timeout, pageService, userService, stateService) {
	var that = this;

	var showDelay = 400; // milliseconds
	var hideDelay = 300; // milliseconds
	var smallPopoverWidth = 400; // pixels
	var largePopoverWidth = 600; // pixels
	var awayFromEdge = 20; // min distance from edge of the screen in pixels

	var mousePageX: number;
	var mousePageY: number;

	var linkTypeIntrasite = 'intrasite';
	var linkTypeUser = 'user';
	var linkTypeText = 'text';

	var popoverScope;
	// The topmost popover element
	var $popoverElement;
	// The topmost anchor element
	var $anchorElement;
	// The stack of previous popover elements
	var popoverElementStack = [];
	// The stack of previous anchor elements
	var anchorElementStack = [];

	var $targetCandidate = undefined;
	var targetCandidateLinkType;
	var createPromise = undefined;
	var removePromise = undefined;
	var anchorHovering = false;
	var popoverHovering = false;

	// Remove all popovers
	var removeAllPopovers = function() {
		while ($popoverElement) {
			removePopover();
		}
	};

	// Remove the popover.
	var removePopover = function() {
		// Remove the popoverElement, and get the next one if there is one.
		if ($popoverElement) {
			popoverScope.$destroy();
			$popoverElement.remove();
		}
		$popoverElement = undefined;
		if (popoverElementStack.length > 0) {
			$popoverElement = popoverElementStack.pop();
		}

		// Get the next anchorElement down.
		$anchorElement = undefined;
		if (anchorElementStack.length > 0) {
			$anchorElement = anchorElementStack.pop();
		}

		$targetCandidate = undefined;
		createPromise = undefined;
		removePromise = undefined;
		anchorHovering = false;
		popoverHovering = false;
	};

	var shutItDown = function() {
		$timeout.cancel(createPromise);
		$timeout.cancel(removePromise);
		removeAllPopovers();
	};

	// Update the timeout timer.
	var updateTimeout = function() {
		if (anchorHovering || $popoverElement.popoverHovering) {
			// Cancel timeout
			$timeout.cancel(removePromise);
			removePromise = undefined;
		} else {
			if (!removePromise) {
				// Start the timer to remove the popover
				removePromise = $timeout(function() {
					while ($popoverElement && !$popoverElement.popoverHovering) {
						removePopover();
					}
				}, hideDelay);
			}
		}
	};

	// Create a new intrasite popover.
	var createPopover = function(event) {
		var $target = $(event.currentTarget);

		// If mouse is in the top part of the screen, show popover down, otherwise up.
		var mouseInTopPart = ((mousePageY - $('body').scrollTop()) / $(window).height()) <= 0.4;
		var direction = mouseInTopPart ? 'down' : 'up';

		var popoverWidth = largePopoverWidth;
		if (targetCandidateLinkType == linkTypeText) {
			popoverWidth = smallPopoverWidth;
		}

		var left = Math.max(0, mousePageX - popoverWidth / 2 - awayFromEdge) + awayFromEdge;
		left = Math.min(left, $('body').width() - popoverWidth - awayFromEdge);
		if (isTouchDevice) left = 0;
		var arrowOffset = mousePageX - left;

		// Create the popover
		popoverScope = $rootScope.$new();
		if ($popoverElement) {
			popoverElementStack.push($popoverElement);
		}
		if (targetCandidateLinkType == linkTypeIntrasite) {
			$popoverElement = $compile('<arb-intrasite-popover page-id=\'' + $target.attr('page-id') +
				'\' direction=\'' + direction + '\' arrow-offset=\'' + arrowOffset +
				'\'></arb-intrasite-popover>')(popoverScope);
		} else if (targetCandidateLinkType == linkTypeUser) {
			$popoverElement = $compile('<arb-user-popover user-id=\'' + $target.attr('user-id') +
				'\' direction=\'' + direction + '\' arrow-offset=\'' + arrowOffset +
				'\'></arb-user-popover>')(popoverScope);
		} else if (targetCandidateLinkType == linkTypeText) {
			// NOTE: it's important to have normal-quotes around encoded HTML
			$popoverElement = $compile('<arb-text-popover encoded-html="' + $target.attr('encoded-html') +
				'" direction=\'' + direction + '\' arrow-offset=\'' + arrowOffset +
				'\'></arb-user-popover>')(popoverScope);
		} else {
			console.error('Unknown link type: ' + targetCandidateLinkType);
		}

		// Set popover properties
		if (mouseInTopPart) {
			var top = mousePageY + parseInt($target.css('font-size'));
			$popoverElement.css('top', top);
		} else {
			var top = mousePageY - parseInt($target.css('font-size'));
			$popoverElement.css('bottom', $('body').height() - top);
		}
		$popoverElement.css('left', left)
			.css('position', '') // IE fix, because it sets position to "relative"
			.width(isTouchDevice ? $('body').width() - 6 : popoverWidth);

		var thisPopoverElement = $popoverElement;
		$popoverElement.on('mouseenter', function(event) {
			thisPopoverElement.popoverHovering = true;
			updateTimeout();
		});
		$popoverElement.on('mouseleave', function(event) {
			thisPopoverElement.popoverHovering = false;
			updateTimeout();
		});

		$('body').append($popoverElement);

		if ($anchorElement) {
			anchorElementStack.push($anchorElement);
		}
		$anchorElement = $target;
		anchorHovering = true;
	};

	var mouseEnterPopoverLink = function(event, linkType) {
		var $target = $(event.currentTarget);
		if ($target.hasClass('red-link')) return;

		// DO allow recursive hover in popovers
		// if ($target.closest('arb-intrasite-popover').length > 0) return;
		// if ($target.closest('arb-user-popover').length > 0) return;
		// if ($target.closest('.md-button').length > 0) return;
		// if ($target.closest('arb-text-popover').length > 0) return;

		if ($anchorElement && $target[0] == $anchorElement[0]) {
			// Hovering over the element we already created a popover for
			anchorHovering = true;
			updateTimeout();
			return;
		}

		if (!$targetCandidate || $target[0] != $targetCandidate[0]) {
			if ($targetCandidate && $target[0] != $targetCandidate[0]) {
				$timeout.cancel(createPromise);
			}

			createPromise = $timeout(createPopover, showDelay, true, event);
			$targetCandidate = $target;
			targetCandidateLinkType = linkType;

			// Prefetch the data
			if (targetCandidateLinkType == linkTypeIntrasite) {
				var pageId = $target.attr('page-id');
				var page = stateService.pageMap[pageId];
				if (!page || Object.keys(page.summaries).length <= 0) {
					pageService.loadIntrasitePopover(pageId);
				}
			} else if (targetCandidateLinkType == linkTypeUser) {
				var userId = $target.attr('user-id');
				var page = stateService.pageMap[userId];
				if (!page || Object.keys(page.summaries).length <= 0) {
					userService.loadUserPopover(userId);
				}
			} else if (targetCandidateLinkType == linkTypeText) {
			} else {
				console.error('Unknown link type: ' + targetCandidateLinkType);
			}
		}
	};

	$('body').on('mouseenter', '.intrasite-link', function(event) {
		mouseEnterPopoverLink(event, linkTypeIntrasite);
	});

	$('body').on('mouseenter', '.user-link', function(event) {
		mouseEnterPopoverLink(event, linkTypeUser);
	});

	$('body').on('mouseenter', '[arb-text-popover-anchor]', function(event) {
		mouseEnterPopoverLink(event, linkTypeText);
	});

	var mouseMovePopoverLink = function(event) {
		mousePageX = event.pageX;
		mousePageY = event.pageY;
	};

	$('body').on('mousemove', '.intrasite-link', function(event) {
		mouseMovePopoverLink(event);
	});

	$('body').on('mousemove', '.user-link', function(event) {
		mouseMovePopoverLink(event);
	});

	$('body').on('mousemove', '[arb-text-popover-anchor]', function(event) {
		mouseMovePopoverLink(event);
	});

	var mouseLeavePopoverLink = function(event) {
		var $target = $(event.currentTarget);
		if ($anchorElement && $target[0] == $anchorElement[0]) {
			// Leaving the element we created a popover for
			anchorHovering = false;
			updateTimeout();
			return;
		}
		if ($targetCandidate && $target[0] == $targetCandidate[0]) {
			// Leaving the element we hovered over for a bit
			$targetCandidate = undefined;
			targetCandidateLinkType = undefined;
			$timeout.cancel(createPromise);
		}
	};

	$('body').on('mouseleave', '.intrasite-link', function(event) {
		mouseLeavePopoverLink(event);
	});

	$('body').on('mouseleave', '.user-link', function(event) {
		mouseLeavePopoverLink(event);
	});

	$('body').on('mouseleave', '[arb-text-popover-anchor]', function(event) {
		mouseLeavePopoverLink(event);
	});

	// On mobile, we want to intercept the click.
	if (isTouchDevice) {
		var touchDeviceLinkClick = function(event, linkType) {
			var $target = $(event.currentTarget);
			if ($target.is($anchorElement)) {
				// User clicked on a link that already has a popover up
				return true;
			}
			if (!$target.is($targetCandidate)) {
				// User clicked on a link that's most likely inside a popover
				return true;
			}
			mouseMovePopoverLink(event);
			return false;
		};

		$('body').on('click', '.intrasite-link', function(event) {
			return touchDeviceLinkClick(event, linkTypeIntrasite);
		});

		$('body').on('click', '.user-link', function(event) {
			return touchDeviceLinkClick(event, linkTypeUser);
		});

		$('body').on('click', '[arb-text-popover-anchor]', function(event) {
			return touchDeviceLinkClick(event, linkTypeText);
		});
	} else {
		// On desktop, clicking the link kills the popover
		$('body').on('click', '.intrasite-link', function(event) {
			shutItDown();
		});

		$('body').on('click', '.user-link', function(event) {
			shutItDown();
		});
	}

	$rootScope.$on('$locationChangeStart', function(event) {
		shutItDown();
	});
});

