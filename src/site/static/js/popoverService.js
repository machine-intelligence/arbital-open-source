"use strict";

// Popover service is used to display the intrasite popover.
app.service("popoverService", function($rootScope, $compile, $timeout, pageService, userService){
	var that = this;

	var showDelay = 400, hideDelay = 300; // milliseconds
	var popoverWidth = 600; // pixels
	var awayFromEdge = 20; // min distance from edge of the screen in pixels

	var mousePageX, mousePageY;

	var linkTypeIntrasite = "intrasite";
	var linkTypeUser = "user";

	var $targetCandidate,
		targetCandidateLinkType,
		createPromise;

	var $popoverElement,
		$currentTarget,
		removePromise,
		anchorHovering,
		popoverHovering;

	// Remove the popover.
	var removePopover = function() {
		if ($popoverElement) {
			$popoverElement.remove();
		}
		$targetCandidate = undefined;
		createPromise = undefined;
		$popoverElement = undefined;
		$currentTarget = undefined;
		removePromise = undefined;
		anchorHovering = false;
		popoverHovering = false;
	};
	removePopover(); // init all the variables

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

		// Create the popover
		if (targetCandidateLinkType == linkTypeIntrasite) {
			$popoverElement = $compile("<arb-intrasite-popover page-id='" + $target.attr("page-id") +
				"'></arb-intrasite-popover>")($rootScope);
		} else if (targetCandidateLinkType == linkTypeUser) {
			$popoverElement = $compile("<arb-user-popover user-id='" + $target.attr("user-id") +
				"'></arb-user-popover>")($rootScope);
		}
		var left = Math.max(0, mousePageX - popoverWidth / 2 - awayFromEdge) + awayFromEdge;
		left = Math.min(left, $("body").width() - popoverWidth - awayFromEdge);
		var top = mousePageY + parseInt($target.css("font-size"));
		$popoverElement.offset({left: left, top: top});
		$popoverElement.width(popoverWidth);
		$popoverElement.on("mouseenter", function(event) {
			popoverHovering = true;
			updateTimeout();
		});
		$popoverElement.on("mouseleave", function(event) {
			popoverHovering = false;
			updateTimeout();
		});
		$("#dynamic-view").append($popoverElement);

		$currentTarget = $target;
		anchorHovering = true;
	};

	$("body").on("mouseenter", ".intrasite-link", function(event) {
		mouseEnterPopoverLink(event, linkTypeIntrasite);
	});

	$("body").on("mouseenter", ".user-link", function(event) {
		mouseEnterPopoverLink(event, linkTypeUser);
	});

	var mouseEnterPopoverLink = function(event, linkType) {
		var $target = $(event.currentTarget);
		if ($target.hasClass("red-link")) return;
		// Don't allow recursive hover in popovers.
		if ($target.closest("arb-intrasite-popover").length > 0) return;
		if ($target.closest("arb-user-popover").length > 0) return;
		if ($currentTarget && $target[0] == $currentTarget[0]) {
			// Hovering over the element we already created a popover for
			anchorHovering = true;
			updateTimeout();
			return;
		}
		
		if (!$targetCandidate) {
			createPromise = $timeout(createPopover, showDelay, true, event);
			$targetCandidate = $target;
			targetCandidateLinkType = linkType;
		} else if ($target[0] != $targetCandidate[0]) {
			$timeout.cancel(createPromise);
			createPromise = $timeout(createPopover, showDelay, true, event);
			$targetCandidate = $target;
			targetCandidateLinkType = linkType;
		}
	};

	$("body").on("mousemove", ".intrasite-link", function(event) {
		mouseMovePopoverLink(event);
	});

	$("body").on("mousemove", ".user-link", function(event) {
		mouseMovePopoverLink(event);
	});

	var mouseMovePopoverLink = function(event) {
		mousePageX = event.pageX;
		mousePageY = event.pageY;
	};

	$("body").on("mouseleave", ".intrasite-link", function(event) {
		mouseLeavePopoverLink(event);
	});

	$("body").on("mouseleave", ".user-link", function(event) {
		mouseLeavePopoverLink(event);
	});

	var mouseLeavePopoverLink = function(event) {
		var $target = $(event.currentTarget);
		if ($currentTarget && $target[0] == $currentTarget[0]) {
			// Leaving the element we created a popover for
			anchorHovering = false;
			updateTimeout();
			return
		}
		if ($targetCandidate && $target[0] == $targetCandidate[0]){
			// Leaving the element we hovered over for a bit
			$targetCandidate = undefined;
			targetCandidateLinkType = undefined;
			$timeout.cancel(createPromise);
		}
	};

	$rootScope.$on("$locationChangeStart", function(event) {
		$timeout.cancel(createPromise);
		$timeout.cancel(removePromise);
		removePopover();
	});
});

