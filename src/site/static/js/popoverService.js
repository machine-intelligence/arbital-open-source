'use strict';

// Popover service is used to display the intrasite popover.
app.service('popoverService', function($rootScope, $compile, $timeout, pageService, userService, stateService) {
	// TODO: enable popovers on mobile again once we can make them good
	if (isTouchDevice) return;

	var that = this;

	var showDelay = 400; // milliseconds
	var hideDelay = 300; // milliseconds
	var popoverWidth = 600; // pixels
	var awayFromEdge = 20; // min distance from edge of the screen in pixels

	var mousePageX;
	var mousePageY;

	var linkTypeIntrasite = 'intrasite';
	var linkTypeUser = 'user';

	var popoverScope;
	var $popoverElement;
	var $currentTarget;
	var $targetCandidate;

	var	targetCandidateLinkType;
	var createPromise;
	var removePromise;
	var anchorHovering;
	var popoverHovering;

	// Remove the popover.
	var removePopover = function() {
		if ($popoverElement) {
			popoverScope.$destroy();
			$popoverElement.remove();
		}
		$targetCandidate = undefined;
		$popoverElement = undefined;
		$currentTarget = undefined;
		createPromise = undefined;
		removePromise = undefined;
		anchorHovering = false;
		popoverHovering = false;
	};
	removePopover(); // init all the variables

	var shutItDown = function() {
		$timeout.cancel(createPromise);
		$timeout.cancel(removePromise);
		removePopover();
	};

	// Update the timeout timer.
	var updateTimeout = function() {
		if (anchorHovering || popoverHovering) {
			// Cancel timeout
			$timeout.cancel(removePromise);
			removePromise = undefined;
		} else {
			if (!removePromise) {
				// Start the timer to remove the popover
				removePromise = $timeout(removePopover, hideDelay);
			}
		}
	};

	// Create a new intrasite popover.
	var createPopover = function(event) {
		var $target = $(event.currentTarget);

		// Delete old popover
		removePopover();

		// If mouse is in the top part of the screen, show popover down, otherwise up.
		var mouseInTopPart = ((mousePageY - $('body').scrollTop()) / $(window).height()) <= 0.4;
		var direction = mouseInTopPart ? 'down' : 'up';

		var left = Math.max(0, mousePageX - popoverWidth / 2 - awayFromEdge) + awayFromEdge;
		left = Math.min(left, $('body').width() - popoverWidth - awayFromEdge);
		if (isTouchDevice) left = 0;
		var arrowOffset = mousePageX - left;

		// Create the popover
		popoverScope = $rootScope.$new();
		if (targetCandidateLinkType == linkTypeIntrasite) {
			$popoverElement = $compile('<arb-intrasite-popover page-id=\'' + $target.attr('page-id') +
				'\' direction=\'' + direction + '\' arrow-offset=\'' + arrowOffset +
				'\'></arb-intrasite-popover>')(popoverScope);
		} else if (targetCandidateLinkType == linkTypeUser) {
			$popoverElement = $compile('<arb-user-popover user-id=\'' + $target.attr('user-id') +
				'\' direction=\'' + direction + '\' arrow-offset=\'' + arrowOffset +
				'\'></arb-user-popover>')(popoverScope);
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
		.width(isTouchDevice ? $('body').width() : popoverWidth)
		.on('mouseenter', function(event) {
			popoverHovering = true;
			updateTimeout();
		})
		.on('mouseleave', function(event) {
			popoverHovering = false;
			updateTimeout();
		});

		$('body').append($popoverElement);
		$currentTarget = $target;
		anchorHovering = true;
	};

	var mouseEnterPopoverLink = function(event, linkType) {
		var $target = $(event.currentTarget);
		if ($target.hasClass('red-link')) return;
		// Don't allow recursive hover in popovers.
		if ($target.closest('arb-intrasite-popover').length > 0) return;
		if ($target.closest('arb-user-popover').length > 0) return;
		if ($target.closest('.md-button').length > 0) return;
		if ($currentTarget && $target[0] == $currentTarget[0]) {
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
				var page = pageService.pageMap[pageId];
				if (!page || Object.keys(page.summaries).length <= 0) {
					pageService.loadIntrasitePopover(pageId);
				}
			} else {
				var userId = $target.attr('user-id');
				var page = pageService.pageMap[userId];
				if (!page || Object.keys(page.summaries).length <= 0) {
					userService.loadUserPopover(userId);
				}
			}
		}
	};

	$('body').on('mouseenter', '.intrasite-link', function(event) {
		mouseEnterPopoverLink(event, linkTypeIntrasite);
	});

	$('body').on('mouseenter', '.user-link', function(event) {
		mouseEnterPopoverLink(event, linkTypeUser);
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

	var mouseLeavePopoverLink = function(event) {
		var $target = $(event.currentTarget);
		if ($currentTarget && $target[0] == $currentTarget[0]) {
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

	// On mobile, we want to intercept the click.
	if (isTouchDevice) {
		var touchDeviceLinkClick = function(event, linkType) {
			var $target = $(event.currentTarget);
			if ($target.is($currentTarget)) {
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

