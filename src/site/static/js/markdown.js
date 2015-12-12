"use strict";

var notEscaped = "(^|\\\\`|\\\\\\[| |>|\n)";
var noParen = "(?=$|[^(])";
var aliasMatch = "([A-Za-z0-9]+\\.?[A-Za-z0-9]*)";
// [vote: alias]
var voteEmbedRegexp = new RegExp(notEscaped + 
		"\\[vote: ?" + aliasMatch + "\\]" + noParen, "g");
// [alias/url text] 
var forwardLinkRegexp = new RegExp(notEscaped + 
		"\\[([^ \\]]+?) ([^\\]]+?)\\]" + noParen, "g");
// [alias]
var simpleLinkRegexp = new RegExp(notEscaped + 
		"\\[" + aliasMatch + "\\]" + noParen, "g");
// [text](alias)
var complexLinkRegexp = new RegExp(notEscaped + 
		"\\[([^\\]]+?)\\]" + // match [Text]
		"\\(" + aliasMatch + "\\)", "g"); // match (Alias)
// [text](url)
var urlLinkRegexp = new RegExp(notEscaped + 
		"\\[([^\\]]+?)\\]" + // match [Text]
		"\\((http://" + RegExp.escape(window.location.host) + "/pages/)" + aliasMatch + "\\)", "g"); // match (Url)
// [@alias]
var atAliasRegexp = new RegExp(notEscaped + 
		"\\[@" + aliasMatch + "\\]" + noParen, "g");

// markdownService provides a constructor you can use to create a markdown converter,
// either for converting markdown to text or editing.
app.service("markdownService", function(pageService, userService){
	// Pass in a pageId to create an editor for that page
	var createConverter = function(pageId, postConversionCallback) {
		// NOTE: not using $location, because we need port number
		var host = window.location.host;
		var converter = Markdown.getSanitizingConverter();

		// Process [todo:text] spans.
		var todoSpanRegexp = new RegExp(notEscaped + 
				"\\[todo: ?([^\\]]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(todoSpanRegexp, function (whole, prefix, alias) {
				return prefix;
			});
		});

		// Process [comment:text] spans.
		var commentSpanRegexp = new RegExp(notEscaped + 
				"\\[comment: ?([^\\]]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(commentSpanRegexp, function (whole, prefix, alias) {
				return prefix;
			});
		});

		// Process [vote:alias] spans.
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(voteEmbedRegexp, function (whole, prefix, alias) {
				return prefix + "[Embedded " + alias + " vote. ](http://" + host + "/pages/" + alias + "/?embedVote=1)";
			});
		});

		// Convert [ text] spans into links.
		var spaceTextRegexp = new RegExp(notEscaped + 
				"\\[ ([^\\]]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(spaceTextRegexp, function (whole, prefix, text) {
				return prefix + "[" + text + "](" + "0" + ")";
			});
		});

		// Convert [alias/url text] spans into links.
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(forwardLinkRegexp, function (whole, prefix, alias, text) {
				return prefix + "[" + text + "](" + alias + ")";
			});
		});

		// Convert [alias] spans into links.
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(simpleLinkRegexp, function (whole, prefix, alias) {
				var page = pageService.pageMap[alias];
				if (page) {
					var url = "http://" + host + "/pages/" + page.pageId;
					return prefix + "[" + page.title + "](" + url + ")";
				} else {
					return prefix + "[" + alias + "](" + alias + ")";
				}
			});
		});

//debugger;
		// Convert [@alias] spans into links.
		converter.hooks.chain("preSpanGamut", function (text) {
//debugger;
			return text.replace(atAliasRegexp, function (whole, prefix, alias) {
//debugger;
				var page = pageService.pageMap[alias];
				if (page) {
					var url = "http://" + host + "/user/" + page.pageId + "/";
					return prefix + "[" + page.title + "](" + url + ")";
				} else {
					var url = "http://" + host + "/user/" + alias + "/";
					return prefix + "[" + alias + "](" + url + ")";
				}
			});
		});
	
		// Convert [Text](Alias) spans into links.
		var aliasRegexp = new RegExp(aliasMatch, "");
		converter.hooks.chain("preSpanGamut", function (text) {
//debugger;
			return text.replace(complexLinkRegexp, function (whole, prefix, text, alias) {
//debugger;
				if (alias.match(aliasRegexp)) {
					var url = "http://" + host + "/pages/" + alias;
					return prefix + "[" + text + "](" + url + ")";
				} else {
					return prefix + "[" + text + "](" + alias + ")";
				}
			});
		});


		if (pageId) {
			// Setup the editor stuff.
			var editor = new Markdown.Editor(converter, pageId);
		}
		
		InitMathjax(converter);

		if (editor) {
			editor.run();

			converter.hooks.chain("postConversion", function (text) {
				if (postConversionCallback) {
					postConversionCallback(function() {
						editor.refreshPreview();
					});
				}
				return text;
			});
		}
		return converter;
	};

	// Process all the links in the give element.
	// If refreshFunc is set, fetch page's we don't have yet, and call that function
	// when one of them is loaded.
	this.processLinks = function($pageText, refreshFunc) {
		// Setup attributes for page links that are within our domain.
		// NOTE: not using $location, because we need port number
		var pageRe = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
			window.location.host + // match the url host part
			"\/pages\/" + aliasMatch + // [1] capture page alias
			"\/?" + // optional ending /
			"(.*)"); // optional other stuff

		// Setup attributes for user links that are within our domain.
		var userRe = new RegExp("^(?:https?:\/\/)?(?:www\.)?" + // match http and www stuff
			window.location.host + // match the url host part
			"\/user\/" + aliasMatch + // [1] capture user alias
			"\/?" + // optional ending /
			"(.*)"); // optional other stuff

		$pageText.find("a").each(function(index, element) {
//debugger;
			var $element = $(element);
			var parts = $element.attr("href").match(pageRe);
			if (parts !== null)	{
				var pageAlias = parts[1];
				
				if (!$element.hasClass("intrasite-link")) {
					$element.addClass("intrasite-link").attr("page-id", pageAlias);
					// Check if we are embedding a vote
					if (parts[2].indexOf("embedVote") > 0) {
						$element.attr("embed-vote-id", pageAlias);
					} else if (pageAlias in pageService.pageMap) {
						if (pageService.pageMap[pageAlias].isDeleted()) {
							// Link to a deleted page.
							$element.addClass("red-link");
						} else {
							// Normal healthy link!
						}
					} else {
						// Mark as red link
						$element.attr("href", $element.attr("href").replace(/pages/, "edit"));
						$element.addClass("red-link");
						if (refreshFunc) {
							pageService.loadTitle(pageAlias, {
								silentFail: true,
								success: function() {
									refreshFunc();
								}
							});
						}
					}
				}
			}

			parts = $element.attr("href").match(userRe);
			if (parts !== null) {
				var userAlias = parts[1];
				
				if (!$element.hasClass("user-link")) {
					$element.addClass("user-link").attr("user-id", userAlias);
					if (userAlias in pageService.pageMap) {
					} else {
						// Mark as red link
						//$element.attr("href", $element.attr("href").replace(/pages/, "edit"));
						$element.addClass("red-link");
						if (refreshFunc) {
							pageService.loadTitle(userAlias, {
								//silentFail: true,
								success: function() {
//debugger;
									refreshFunc();
								}
							});
						}
					}
				}
			}
		});

		$pageText.find("h1").addClass("md-display-2");
		$pageText.find("h2").addClass("md-display-1");
		$pageText.find("h3").addClass("md-headline");
	};

	this.createConverter = function() {
		return createConverter();
	};

	this.createEditConverter = function(pageId, postConversionCallback) {
		return createConverter(pageId, postConversionCallback);
	};
});

// Directive for rendering markdown text.
app.directive("arbMarkdown", function ($compile, $timeout, pageService, markdownService) {
	return {
		scope: {
			pageId: "@",
			useSummary: "@",
		},
		controller: function($scope) {
			$scope.page = pageService.pageMap[$scope.pageId];
		},
		link: function(scope, element, attrs) {
			element.addClass("markdown-text reveal-after-render");

			// Remove the class manually in case the text is empty
			$timeout(function() {
				element.removeClass("reveal-after-render");
			});

			// Convert page text to html.
//debugger;
			var converter = markdownService.createConverter();
			var html = converter.makeHtml(scope.useSummary ? scope.page.summary : scope.page.text);
			var $pageText = element;
			$pageText.html(html);
			window.setTimeout(function() {
				MathJax.Hub.Queue(["Typeset", MathJax.Hub, $pageText.get(0)]);
			}, 100);
//debugger;
			markdownService.processLinks($pageText);
		},
	};
});
