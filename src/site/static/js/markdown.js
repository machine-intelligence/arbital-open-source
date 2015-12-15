"use strict";

var notEscaped = "(^|\\\\`|\\\\\\[|[ *(>\n])";
var noParen = "(?=$|[^(])";
var aliasMatch = "([A-Za-z0-9_]+\\.?[A-Za-z0-9_]*)";
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
	// Store an array of page aliases that failed to load, so that we don't keep trying to reload them
	var failedPageAliases = {};

	// Pass in a pageId to create an editor for that page
	var createConverter = function(pageId, postConversionCallback) {
		// NOTE: not using $location, because we need port number
		var host = window.location.host;
		var converter = Markdown.getSanitizingConverter();

		// Process [todo:text] spans.
		var todoSpanRegexp = new RegExp(notEscaped + 
				"\\[todo: ?([^\\]]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(todoSpanRegexp, function (whole, prefix, text) {
				return prefix;
			});
		});

		// Process [comment:text] spans.
		var commentSpanRegexp = new RegExp(notEscaped + 
				"\\[comment: ?([^\\]]+?)\\]" + noParen, "g");
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(commentSpanRegexp, function (whole, prefix, text) {
				return prefix;
			});
		});

		// Process [summary:markdown] spans.
		var summarySpanRegexp = new RegExp("^\\[summary: ?([\\s\\S]+?)\\] *(?=\Z|\n\Z|\n\n)", "gm");
		converter.hooks.chain("preBlockGamut", function (text, runBlockGamut) {
			return text.replace(summarySpanRegexp, function (whole, summary) {
				if (pageId) {
					return runBlockGamut("---\n\n**Summary:** " + summary + "\n\n---");
				} else {
					return runBlockGamut("");
				}
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

		// Convert [@alias] spans into links.
		converter.hooks.chain("preSpanGamut", function (text) {
			return text.replace(atAliasRegexp, function (whole, prefix, alias) {
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
			return text.replace(complexLinkRegexp, function (whole, prefix, text, alias) {
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
			InitMathjax(converter, editor, pageId);
			editor.run();

			converter.hooks.chain("postConversion", function (text) {
				if (postConversionCallback) {
					postConversionCallback(function() {
						editor.refreshPreview();
					});
				}
				return text;
			});
		} else {
			InitMathjax(converter);
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
						if (refreshFunc && !(pageAlias in failedPageAliases) ) {
							pageService.loadTitle(pageAlias, {
								silentFail: true,
								success: function() {
									if (pageAlias in pageService.pageMap) {
										refreshFunc();
									} else {
										failedPageAliases[pageAlias] = true;
									}
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
						$element.addClass("red-link");
						if (refreshFunc && !(userAlias in failedPageAliases) ) {
							pageService.loadTitle(userAlias, {
								silentFail: true,
								success: function() {
									if (userAlias in pageService.pageMap) {
										refreshFunc();
									} else {
										failedPageAliases[userAlias] = true;
									}
								}
							});
						}
					}
				}
			}
		});
	};

	this.createConverter = function() {
		return createConverter();
	};

	this.createEditConverter = function(pageId, postConversionCallback) {
		failedPageAliases = {};
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
			var converter = markdownService.createConverter();
			var html = converter.makeHtml(scope.useSummary ? scope.page.summary : scope.page.text);
			var $pageText = element;
			$pageText.html(html);
			window.setTimeout(function() {
				MathJax.Hub.Queue(["Typeset", MathJax.Hub, $pageText.get(0)]);
			}, 100);
			markdownService.processLinks($pageText);
		},
	};
});
