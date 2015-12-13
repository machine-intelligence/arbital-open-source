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
			autocompleteService.setupParentsAutocomplete($input, function(event, ui) {
				element.find(".modal-content").submit();
				return true;
			});
			// Set up search for new link modal
			autocompleteService.setAutocompleteRendering($input, scope);
		},
	};
});

// intrasitePopover contains the popover body html.
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
			arbMarkdown.init(false, scope.pageId, scope.page.summary, element.find(".intrasite-popover-body"), pageService, userService);
		},
	};
});

// pageTitle displays page's title with optional meta info.
app.directive("arbPageTitle", function(pageService, userService) {
	return {
		templateUrl: "/static/html/pageTitle.html",
		scope: {
			pageId: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
		},
	};
});

// likesPageTitle displays likes span followed by page's title span.
app.directive("arbLikesPageTitle", function(pageService, userService) {
	return {
		templateUrl: "/static/html/likesPageTitle.html",
		scope: {
			pageId: "@",
			showClickbait: "@",
			showRedLinkCount: "@",
			showQuickEditLink: "@",
			showCreatedAt: "@",
			isSearchResult: "@",
			isSupersized: "@",
		},
		link: function(scope, element, attrs) {
			scope.pageService = pageService;
			scope.userService = userService;
			scope.page = pageService.pageMap[scope.pageId];
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
